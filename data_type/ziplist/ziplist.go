package ziplist

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"reflect"
	"unsafe"

	rs "github.com/WANGgbin/tiny_redis/data_type/redis_string"
	"github.com/WANGgbin/tiny_redis/utils"
)

// ziplist 本质上是一块连续的内存，在数据量较少也及时间复杂度在可容忍范围内，相较与链表、哈系表等
// 因为没有指针，节省了内存空间。
// note: 所有数据以大端形式存储

const (
	EncodingMask    = 0xc0
	EncodingInt8    = 0xc0 // 1100 0000
	EncodingInt16   = 0xd0 // 1101 0000
	EncodingInt24   = 0xe0 // 1110 0000
	EncodingInt32   = 0xfd // 1111 1101
	EncodingInt64   = 0xfe // 1111 1110
	EncodingIMMMax  = 0xfc // 1111 0000 - 1111 1100 (0 - 12)
	EncodingIMMMin  = 0xf0
	Tail            = 0xff
	Prev5BytesBegin = 0xfe

	OneByteBinLength  = 0x3f
	TwoByteBinLength  = 0x3fff
	FourByteBinLength = 0xffffffff
)

var (
	PointedToEndErr = errors.New("points to the end of ziplist")
	PointerIsNilErr = errors.New("pointer is nil")
)

type ZipList struct {
	content []byte
}

type dataType uint8

const (
	binData dataType = iota
	intData
)

// 定义 zlEntry 是方便解析/反解析操作
type zlEntry struct {
	offset          uint32
	prevFieldLength uint32 // 指的是 prev 字段的长度，而不是前一个节点的长度
	prevVal         []byte
	encodingLength  uint32
	encodingVal     []byte
	dataLength      uint32
	data            []byte
	encodingType    dataType
}

func (entry *zlEntry) getLength() uint32 {
	return entry.prevFieldLength + entry.encodingLength + entry.dataLength
}

func (entry *zlEntry) getPrevLength() uint32 {
	if entry.prevVal[0] == Prev5BytesBegin {
		return binary.BigEndian.Uint32(entry.prevVal[1:])
	}
	return uint32(entry.prevVal[0])
}

// -----------------------------------------------------------------------------
// | 字节数量(4Byte) | 尾结点偏移量(4Byte) | 节点个数(2Byte) | entry1 | ... | zlend |
// -----------------------------------------------------------------------------

func CreateZipList(elements ...interface{}) (*ZipList, error) {
	zl := InitZipList()
	for _, elem := range elements {
		err := zl.ZipListPush(elem)
		if err != nil {
			log.Printf("When push element:%v, error happened: %v\n", elem, err)
			return nil, err
		}
	}
	return zl, nil
}

func InitZipList() *ZipList {
	initLen := 4 + 4 + 2 + 2 + 1
	zipList := &ZipList{
		content: make([]byte, initLen),
	}

	zipList.content[initLen-1] = Tail
	binary.BigEndian.PutUint32(zipList.content[:4], 13)
	binary.BigEndian.PutUint32(zipList.content[4:8], 10)

	// 同样为了方便操作，加入一个头节点():0x00 0xf0，该节点不计入节点个数
	zipList.content[10] = 0x00
	zipList.content[11] = 0xf0

	return zipList
}

func (zl *ZipList) getLength() uint32 {
	return binary.BigEndian.Uint32(zl.content[:4])
}

func (zl *ZipList) getOffset() uint32 {
	return binary.BigEndian.Uint32(zl.content[4:8])
}

// ZipListPush pushes an entry at the end of z.
func (zl *ZipList) ZipListPush(val interface{}) error {
	entry, err := zl.convertValToEntry(val)
	if err != nil {
		return err
	}

	// 填充 prevVal & prevFieldLength
	position, _ := opPointer(&zl.content[0], zl.getOffset(), Plus)
	zl.fillPrev(position, entry)
	zl.insertEntry(position, entry)

	return nil
}

// InsertEntry inserts an entry after position.
func (zl *ZipList) ZipListInsert(position *byte, val interface{}) error {
	entry, err := zl.convertValToEntry(val)
	if err != nil {
		return err
	}

	zl.fillPrev(position, entry)
	zl.insertEntry(position, entry)

	return nil
}

func (zl *ZipList) convertValToEntry(val interface{}) (*zlEntry, error) {
	var entry *zlEntry
	var err error
	switch val := val.(type) {
	case nil:
		return nil, errors.New("Val to push is nil")
	case int64:
		entry = zl.transIntToEntry(val)
	case *rs.RedisString:
		entry, err = zl.transBinToEntry(val)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(fmt.Sprintf("Unknown val to push: %v", val))
	}
	return entry, nil
}

func (zl *ZipList) fillPrev(position *byte, entry *zlEntry) {
	prevEntry, _ := zl.decodeBytesToEntry(position)
	prevLength := prevEntry.getLength()

	if prevLength <= 0xfd {
		entry.prevFieldLength = 1
		entry.prevVal = []byte{byte(prevLength)}
	} else {
		entry.prevFieldLength = 5
		entry.prevVal = make([]byte, 5)
		entry.prevVal[0] = 0xfe
		binary.BigEndian.PutUint32(entry.prevVal[1:], prevLength)
	}
	entry.offset = prevEntry.offset + prevLength
}

func (zl *ZipList) insertEntry(position *byte, entry *zlEntry) {
	nextPosition, _ := opPointer(position, entry.getPrevLength(), Plus)
	extendBytes := zl.copyContentForInsert(entry, nextPosition)

	// 调整总长度
	totalBytes := uint32(len(zl.content))
	binary.BigEndian.PutUint32(zl.content[:4], totalBytes)

	// 调整尾部节点偏移量
	offset := binary.BigEndian.Uint32(zl.content[4:8])
	if *nextPosition == 0xff {
		offset += entry.getPrevLength()
	} else {
		offset += entry.getLength() + extendBytes
	}

	binary.BigEndian.PutUint32(zl.content[4:8], offset)

	// 调整结点个数
	count := binary.BigEndian.Uint16(zl.content[8:10])
	count += 1
	binary.BigEndian.PutUint16(zl.content[8:10], count)
}

func (zl *ZipList) copyContentForInsert(entry *zlEntry, nextPosition *byte) uint32 {
	entryBytes := encodeEntryToBytes(entry)
	toExtendEntries, extendBytes := zl.calcBytesToExtend(entry.getLength(), nextPosition)
	newContent := make([]byte, len(zl.content)+len(entryBytes)+int(extendBytes))
	// 填充头部不变部分
	copy(newContent, zl.content[:entry.offset])
	
	// 填充新节点
	copy(newContent[entry.offset:], entryBytes)

	// 填充长度更新节点
	prevLen := entry.getLength() - 4
	curOffset := entry.offset + entry.getLength()
	originCurOffset := entry.offset

	for _, toExtendEntry := range toExtendEntries {
		copy(newContent[curOffset+5:], toExtendEntry.encodingVal)
		copy(newContent[curOffset+5+toExtendEntry.encodingLength:], toExtendEntry.data)
		newContent[curOffset] = 0xfe
		prevLen = prevLen + 4
		binary.BigEndian.PutUint32(newContent[curOffset+1:curOffset+5], prevLen)

		prevLen = toExtendEntry.getLength()
		curOffset += 5 + toExtendEntry.encodingLength + toExtendEntry.dataLength
		originCurOffset += toExtendEntry.getLength()
	}

	// 填充尾部不变部分
	copy(newContent[curOffset:], zl.content[originCurOffset:])

	// 修改最后一个调整结点的下一个结点的 prevVal
	if zl.content[originCurOffset] != 0xff {
		if zl.content[originCurOffset] == 0xfe {
			binary.BigEndian.PutUint32(newContent[curOffset+1:curOffset+5], 4 + prevLen)
		} else {
			newContent[curOffset] = byte(4 + prevLen)
		}
	}

	zl.content = newContent

	return extendBytes
}

func (zl *ZipList) calcBytesToExtend(prevLen uint32, p *byte) ([]*zlEntry, uint32) {
	var toExtendEntries []*zlEntry
	extendBytes := uint32(0)

	for {
		if *p == Tail || !zl.whetherToExtend(prevLen, p) {
			return toExtendEntries, extendBytes
		}

		extendBytes += 4

		entry, _ := zl.decodeBytesToEntry(p)
		toExtendEntries = append(toExtendEntries, entry)
		prevLen = entry.getLength() + 4
		p, _ = opPointer(p, entry.getLength(), Plus)
	}
}

func (zl *ZipList) whetherToExtend(prevLen uint32, p *byte) bool {
	if *p != 0xfe && prevLen > 0xfd {
		return true
	}

	return false
}

func (zl *ZipList) ZipListDelete(val interface{}) (bool, error) {
	findEntry, err := zl.findElem(val)
	if err != nil {
		return false, err
	}

	if findEntry == nil {
		return false, nil
	}

	zl.deleteEntry(findEntry)

	return true, nil
}

func (zl *ZipList) findElem(val interface{}) (*zlEntry, error) {
	targetEntry, err := zl.convertValToEntry(val)
	if err != nil {
		return nil, err
	}

	startIndex := &zl.content[12]
	for {
		if *startIndex == Tail {
			return nil, nil
		}

		curEntry, err := zl.decodeBytesToEntry(startIndex)
		if err != nil {
			return nil, err
		}

		if reflect.DeepEqual(targetEntry.data, curEntry.data) {
			return curEntry, nil
		}

		startIndex, _ = opPointer(startIndex, curEntry.getLength(), Plus)
	}
}

func (zl *ZipList) deleteEntry(entry *zlEntry) {
	prevLen := entry.getPrevLength()
	position, _ := opPointer(&zl.content[0], entry.offset + entry.getLength(), Plus)
	toShrinkEntries, toShrinkBytes := zl.calcBytesToShrink(prevLen, position)

	newContent := make([]byte, len(zl.content) - int(entry.getLength()) - int(toShrinkBytes))

	copy(newContent, zl.content[:entry.offset])

	prevLen += 4
	curOffset := entry.offset
	originOffset := entry.offset + entry.getLength()

	for _, toShrinkEntry := range toShrinkEntries {
		copy(newContent[curOffset + 1:], toShrinkEntry.encodingVal)
		copy(newContent[curOffset + 1 + toShrinkEntry.encodingLength:], toShrinkEntry.data)
		newContent[curOffset] = byte(prevLen - 4)

		prevLen = toShrinkEntry.getLength()
		curOffset += prevLen - 4
		originOffset += prevLen
	}

	copy(newContent[curOffset:], zl.content[originOffset:])
	if zl.content[originOffset] != Tail {
		if zl.content[originOffset] == 0xfe {
			binary.BigEndian.PutUint32(newContent[curOffset + 1: curOffset + 5], prevLen - 4)
		} else {
			newContent[curOffset] = byte(prevLen - 4)
		}
	}

	zl.content = newContent

	// 调整总长度
	totalBytes := uint32(len(zl.content))
	binary.BigEndian.PutUint32(zl.content[:4], totalBytes)

	// 调整尾部节点偏移量
	offset := binary.BigEndian.Uint32(zl.content[4:8])
	if *position == 0xff {
		offset -= entry.getPrevLength()
	} else {
		offset -= entry.getLength() + toShrinkBytes
	}

	binary.BigEndian.PutUint32(zl.content[4:8], offset)

	// 调整结点个数
	count := binary.BigEndian.Uint16(zl.content[8:10])
	count -= 1
	binary.BigEndian.PutUint16(zl.content[8:10], count)
}

func (zl *ZipList) calcBytesToShrink(prevLen uint32, p *byte) ([]*zlEntry, uint32) {
	var toShrinkEntries []*zlEntry
	shrinkBytes := uint32(0)

	for {
		if *p == 0xff || !zl.whetherToShrink(prevLen, p) {
			return toShrinkEntries, shrinkBytes
		}

		shrinkBytes += 4
		entry, _ := zl.decodeBytesToEntry(p)
		toShrinkEntries = append(toShrinkEntries, entry)
		prevLen = entry.getLength() - 4
		p, _ = opPointer(p, entry.getLength(), Plus)
	}
}

func (zl *ZipList) whetherToShrink(prevLen uint32, p *byte) bool {
	if *p == 0xfe && prevLen <= 0xfd {
		return true
	}

	return false
}

func (zl *ZipList) decodeBytesToEntry(p *byte) (*zlEntry, error) {
	if p == nil {
		return nil, PointerIsNilErr
	}

	if *p == Tail {
		return nil, PointedToEndErr
	}
	entry := new(zlEntry)
	zl.decodeFieldPrev(p, entry)

	afterMask := zl.content[entry.offset+entry.prevFieldLength] & EncodingMask
	if afterMask == 0xc0 {
		zl.decodeIntEncoding(entry)
	} else {
		zl.decodeBinEncoding(entry)
	}

	return entry, nil
}

func (zl *ZipList) decodeFieldPrev(p *byte, entry *zlEntry) {
	entry.offset = uint32(uintptr(unsafe.Pointer(p)) - uintptr(unsafe.Pointer(&zl.content[0])))
	if *p == Prev5BytesBegin {
		entry.prevFieldLength = 5
		entry.prevVal = zl.content[entry.offset : entry.offset+5]
	} else {
		entry.prevFieldLength = 1
		entry.prevVal = zl.content[entry.offset : entry.offset+1]
	}
}

// decodeBinEncoding parses bin data
func (zl *ZipList) decodeBinEncoding(entry *zlEntry) {
	entry.encodingType = binData
	afterMask := zl.content[entry.offset+entry.prevFieldLength] & EncodingMask
	if afterMask == 0x00 {
		entry.encodingLength = 1
		entry.dataLength = uint32(zl.content[entry.offset+entry.prevFieldLength] & 0x3f)
	} else if afterMask == 0x01 {
		entry.encodingLength = 2
		tmp := make([]byte, 2)
		copy(tmp, zl.content[entry.offset+entry.prevFieldLength:])
		tmp[0] &= 0x3f
		entry.dataLength = uint32(binary.BigEndian.Uint16(tmp))
	} else {
		entry.encodingLength = 5
		entry.dataLength = binary.BigEndian.Uint32(zl.content[entry.offset+entry.prevFieldLength+1 : entry.offset+entry.prevFieldLength+5])
	}

	entry.encodingVal = zl.content[entry.offset+entry.prevFieldLength : entry.offset+entry.prevFieldLength+entry.encodingLength]
	entry.data = zl.content[entry.offset+entry.prevFieldLength+entry.encodingLength : entry.offset+entry.prevFieldLength+entry.encodingLength+entry.dataLength]
}

// decodeIntEncoding parses int data.
func (zl *ZipList) decodeIntEncoding(entry *zlEntry) {
	entry.encodingLength = 1
	entry.encodingVal = zl.content[entry.offset+entry.prevFieldLength : entry.offset+entry.prevFieldLength+entry.encodingLength]
	entry.encodingType = intData

	switch entry.encodingVal[0] {
	case EncodingInt8:
		entry.dataLength = 1
	case EncodingInt16:
		entry.dataLength = 2
		// entry.dataInt = int64(binary.BigEndian.Uint16(zl.content[entry.offset+entry.prevFieldLength+entry.encodingLength : entry.offset+entry.prevFieldLength+entry.encodingLength+2]))
	case EncodingInt24:
		entry.dataLength = 3
		// entry.dataInt = int64(BigEndianUint24(zl.content[entry.offset+entry.prevFieldLength+entry.encodingLength : entry.offset+entry.prevFieldLength+entry.encodingLength+3]))
	case EncodingInt32:
		entry.dataLength = 4
		// entry.dataInt = int64(binary.BigEndian.Uint32(zl.content[entry.offset+entry.prevFieldLength+entry.encodingLength : entry.offset+entry.prevFieldLength+entry.encodingLength+4]))
	case EncodingInt64:
		entry.dataLength = 8
		// entry.dataInt = int64(binary.BigEndian.Uint64(zl.content[entry.offset+entry.prevFieldLength+entry.encodingLength : entry.offset+entry.prevFieldLength+entry.encodingLength+8]))
	default:
		entry.dataLength = 0
		// entry.dataInt = int64(zl.content[entry.offset+entry.prevFieldLength] & 0x0f)
	}

	entry.data = zl.content[entry.offset+entry.prevFieldLength+entry.encodingLength : entry.offset+entry.prevFieldLength+entry.encodingLength+entry.dataLength]

}

func (zl *ZipList) transIntToEntry(num int64) *zlEntry {
	entry := &zlEntry{
		encodingType:   intData,
		encodingVal:    make([]byte, 1),
		encodingLength: 1,
	}

	if num >= 0 && num <= 12 {
		entry.dataLength = 0
		entry.encodingVal[0] = byte(EncodingIMMMin + num)
	} else if num >= int64(utils.INT8_MIN) && num <= int64(utils.INT8_MAX) {
		entry.dataLength = 1
		entry.data = []byte{byte(num)}
		entry.encodingVal[0] = EncodingInt8
	} else if num >= int64(utils.INT16_MIN) && num <= int64(utils.INT16_MAX) {
		entry.dataLength = 2
		entry.data = make([]byte, 2)
		binary.BigEndian.PutUint16(entry.data, uint16(num))
		entry.encodingVal[0] = EncodingInt16
	} else if num >= int64(utils.INT24_MIN) && num <= int64(utils.INT24_MAX) {
		entry.dataLength = 3
		entry.data = make([]byte, 3)
		BigEndianPutUint24(entry.data, uint32(num))
		entry.encodingVal[0] = EncodingInt24
	} else if num >= int64(utils.INT32_MIN) && num <= int64(utils.INT32_MAX) {
		entry.dataLength = 4
		entry.data = make([]byte, 4)
		binary.BigEndian.PutUint32(entry.data, uint32(num))
		entry.encodingVal[0] = EncodingInt32
	} else {
		entry.dataLength = 8
		entry.data = make([]byte, 8)
		binary.BigEndian.PutUint64(entry.data, uint64(num))
		entry.encodingVal[0] = EncodingInt64
	}

	return entry
}

func (zl *ZipList) transBinToEntry(str *rs.RedisString) (*zlEntry, error) {
	entry := new(zlEntry)
	if str == nil {
		return nil, PointerIsNilErr
	}

	// 注意，这里是浅拷贝，后续可能需要使用 copy
	entry.dataLength = uint32(len(str.Content))
	entry.data = str.Content
	if entry.dataLength <= OneByteBinLength {
		entry.encodingLength = 1
		entry.encodingVal = []byte{byte(entry.dataLength)}
	} else if entry.dataLength <= TwoByteBinLength {
		entry.encodingLength = 2
		encode := make([]byte, 2)
		binary.BigEndian.PutUint16(encode, uint16(entry.dataLength))
		encode[0] |= byte(0x40)
		entry.encodingVal = encode
	} else if entry.dataLength <= uint32(FourByteBinLength) {
		entry.encodingLength = 5
		encode := make([]byte, 5)
		binary.BigEndian.PutUint32(encode[1:], uint32(entry.dataLength))
		encode[0] = byte(0x80)
		entry.encodingVal = encode
	} else {
		return nil, errors.New(fmt.Sprintf("The length of str: is %d which is bigger than limit: %d", entry.dataLength, FourByteBinLength))
	}

	return entry, nil
}

func encodeEntryToBytes(entry *zlEntry) []byte {
	bytes := make([]byte, entry.prevFieldLength+entry.encodingLength+entry.dataLength)
	copy(bytes[:entry.prevFieldLength], entry.prevVal)
	copy(bytes[entry.prevFieldLength:], entry.encodingVal)
	copy(bytes[entry.prevFieldLength+entry.encodingLength:], entry.data)

	return bytes
}

type Op uint8

const (
	Plus  Op = 0
	Minus Op = 1
)

var (
	UnknownPointerOpErr = errors.New("Unknown pointer operation")
)

func opPointer(src *byte, offset uint32, op Op) (*byte, error) {
	if op == Plus {
		return (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(src)) + uintptr(offset))), nil
	} else if op == Minus {
		return (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(src)) - uintptr(offset))), nil
	} else {
		return nil, UnknownPointerOpErr
	}
}

func BigEndianUint24(bytes []byte) uint32 {
	_ = bytes[2]
	return uint32(bytes[0])<<16 | uint32(bytes[1])<<8 | uint32(bytes[2])
}

func BigEndianPutUint24(bytes []byte, num uint32) {
	_ = bytes[2]
	bytes[0] = byte(num >> 16)
	bytes[1] = byte(num >> 8)
	bytes[2] = byte(num)
}
