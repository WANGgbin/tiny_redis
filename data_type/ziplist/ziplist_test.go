package ziplist

import (
	"encoding/binary"
	"encoding/hex"
	"reflect"
	"testing"

	rs "github.com/WANGgbin/tiny_redis/data_type/redis_string"
	"github.com/WANGgbin/tiny_redis/utils"
)

func TestCreateZipList(t *testing.T) {
	type args struct {
		elements []interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *ZipList
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "case1",
			args: args{
				elements: []interface{}{
					int64(0),
					&rs.RedisString{
						Content: []byte{},
					},
				},
			},
			want: &ZipList{
				content: []byte{0x00, 0x00, 0x00, 0x11,
					0x00, 0x00, 0x00, 0x0e,
					0x00, 0x02,
					0x00, 0xf0,
					0x02, 0xf0,
					0x02, 0x00,
					0xff,
				},
			},
		},
		{
			name: "case2",
			args: args{
				elements: []interface{}{
					int64(13),
					&rs.RedisString{
						Content: []byte{0x11, 0x12},
					},
				},
			},
			want: &ZipList{
				content: []byte{0x00, 0x00, 0x00, 0x14,
					0x00, 0x00, 0x00, 0x0f,
					0x00, 0x02,
					0x00, 0xf0,
					0x02, 0xc0, 0x0d,
					0x03, 0x02, 0x11, 0x12,
					0xff,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateZipList(tt.args.elements...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateZipList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateZipList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitZipList(t *testing.T) {
	tests := []struct {
		name string
		want *ZipList
	}{
		// TODO: Add test cases.
		{
			name: "init",
			want: &ZipList{
				content: []byte{0x00, 0x00, 0x00, 0x0d,
					0x00, 0x00, 0x00, 0x0a,
					0x00, 0x00,
					0x00, 0xf0,
					0xff,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitZipList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitZipList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZipList_ZipListPush(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		val interface{}
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		resultFields fields
		wantErr      bool
	}{
		// TODO: Add test cases.
		{
			name: "push1",
			fields: fields{
				content: []byte{0x00, 0x00, 0x00, 0x0d,
					0x00, 0x00, 0x00, 0x0a,
					0x00, 0x00,
					0x00, 0xf0,
					0xff},
			},
			args: args{
				val: int64(1),
			},
			resultFields: fields{
				content: []byte{0x00, 0x00, 0x00, 0x0f,
					0x00, 0x00, 0x00, 0x0c,
					0x00, 0x01,
					0x00, 0xf0,
					0x02, 0xf1,
					0xff},
			},
		},
		{
			name: "push13",
			fields: fields{
				content: []byte{0x00, 0x00, 0x00, 0x0d,
					0x00, 0x00, 0x00, 0x0a,
					0x00, 0x00,
					0x00, 0xf0,
					0xff},
			},
			args: args{
				val: int64(13),
			},
			resultFields: fields{
				content: []byte{0x00, 0x00, 0x00, 0x10,
					0x00, 0x00, 0x00, 0x0c,
					0x00, 0x01,
					0x00, 0xf0,
					0x02, 0xc0, 0x0d,
					0xff},
			},
		},
		{
			name: "push0bytes",
			fields: fields{
				content: []byte{0x00, 0x00, 0x00, 0x0d,
					0x00, 0x00, 0x00, 0x0a,
					0x00, 0x00,
					0x00, 0xf0,
					0xff},
			},
			args: args{
				val: &rs.RedisString{
					Content: []byte{},
				},
			},
			resultFields: fields{
				[]byte{0x00, 0x00, 0x00, 0x0f,
					0x00, 0x00, 0x00, 0x0c,
					0x00, 0x01,
					0x00, 0xf0,
					0x02, 0x00,
					0xff},
			},
		},
		{
			name: "push1bytes",
			fields: fields{
				content: []byte{0x00, 0x00, 0x00, 0x0d,
					0x00, 0x00, 0x00, 0x0a,
					0x00, 0x00,
					0x00, 0xf0,
					0xff},
			},
			args: args{
				val: &rs.RedisString{
					Content: []byte{0xff},
				},
			},
			resultFields: fields{
				[]byte{0x00, 0x00, 0x00, 0x10,
					0x00, 0x00, 0x00, 0x0c,
					0x00, 0x01,
					0x00, 0xf0,
					0x02, 0x01, 0xff,
					0xff},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			wantZl := &ZipList{
				content: tt.resultFields.content,
			}

			if err := zl.ZipListPush(tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("ZipList.ZipListPush() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(zl, wantZl) {
				t.Errorf("ZipList.ZipListPush() zl = %v, wantZl = %v", zl, wantZl)
			}
		})
	}
}

func TestZipList_decodeBytesToEntry(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		p *byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *zlEntry
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			got, err := zl.decodeBytesToEntry(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("ZipList.decodeBytesToEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ZipList.decodeBytesToEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZipList_decodeBinEncoding(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		entry *zlEntry
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			zl.decodeBinEncoding(tt.args.entry)
		})
	}
}

func TestZipList_decodeIntEncoding(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		entry *zlEntry
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			zl.decodeIntEncoding(tt.args.entry)
		})
	}
}

func TestZipList_transIntToEntry(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		num int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *zlEntry
	}{
		// TODO: Add test cases.
		{
			name: "12",
			args: args{
				num: 12,
			},
			want: &zlEntry{
				encodingLength: 1,
				encodingVal:    []byte{EncodingIMMMax},
				dataLength:     0,
				encodingType:   intData,
			},
		},
		{
			name: "0",
			args: args{
				num: 0,
			},
			want: &zlEntry{
				encodingLength: 1,
				encodingVal:    []byte{EncodingIMMMin},
				dataLength:     0,
				encodingType:   intData,
			},
		},
		{
			name: "INT8_MAX",
			args: args{
				num: int64(utils.INT8_MAX),
			},
			want: &zlEntry{
				encodingLength: 1,
				encodingVal:    []byte{EncodingInt8},
				dataLength:     1,
				data:           []byte{0x7f},
				encodingType:   intData,
			},
		},
		{
			name: "INT8_MIN",
			args: args{
				num: int64(utils.INT8_MIN),
			},
			want: &zlEntry{
				encodingLength: 1,
				encodingVal:    []byte{EncodingInt8},
				dataLength:     1,
				data:           []byte{0x80},
				encodingType:   intData,
			},
		},
		{
			name: "INT16_MAX",
			args: args{
				num: int64(utils.INT16_MAX),
			},
			want: &zlEntry{
				encodingLength: 1,
				encodingVal:    []byte{EncodingInt16},
				dataLength:     2,
				data:           []byte{0x7f, 0xff},
				encodingType:   intData,
			},
		},
		{
			name: "INT16_MIN",
			args: args{
				num: int64(utils.INT16_MIN),
			},
			want: &zlEntry{
				encodingLength: 1,
				encodingVal:    []byte{EncodingInt16},
				dataLength:     2,
				data:           []byte{0x80, 0x00},
				encodingType:   intData,
			},
		},
		{
			name: "INT24_MAX",
			args: args{
				num: int64(utils.INT24_MAX),
			},
			want: &zlEntry{
				encodingLength: 1,
				encodingVal:    []byte{EncodingInt24},
				dataLength:     3,
				data:           []byte{0x7f, 0xff, 0xff},
				encodingType:   intData,
			},
		},
		{
			name: "INT24_MIN",
			args: args{
				num: int64(utils.INT24_MIN),
			},
			want: &zlEntry{
				encodingLength: 1,
				encodingVal:    []byte{EncodingInt24},
				dataLength:     3,
				data:           []byte{0x80, 0x00, 0x00},
				encodingType:   intData,
			},
		},
		{
			name: "INT32_MAX",
			args: args{
				num: int64(utils.INT32_MAX),
			},
			want: &zlEntry{
				encodingLength: 1,
				encodingVal:    []byte{EncodingInt32},
				dataLength:     4,
				data:           []byte{0x7f, 0xff, 0xff, 0xff},
				encodingType:   intData,
			},
		},
		{
			name: "INT32_MIN",
			args: args{
				num: int64(utils.INT32_MIN),
			},
			want: &zlEntry{
				encodingLength: 1,
				encodingVal:    []byte{EncodingInt32},
				dataLength:     4,
				data:           []byte{0x80, 0x00, 0x00, 0x00},
				encodingType:   intData,
			},
		},
		{
			name: "INT64_MAX",
			args: args{
				num: int64(utils.INT64_MAX),
			},
			want: &zlEntry{
				encodingLength: 1,
				encodingVal:    []byte{EncodingInt64},
				dataLength:     8,
				data:           []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				encodingType:   intData,
			},
		},
		{
			name: "INT64_MIN",
			args: args{
				num: int64(utils.INT64_MIN),
			},
			want: &zlEntry{
				encodingLength: 1,
				encodingVal:    []byte{EncodingInt64},
				dataLength:     8,
				data:           []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				encodingType:   intData,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			if got := zl.transIntToEntry(tt.args.num); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ZipList.transIntToEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZipList_transBinToEntry(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		str *rs.RedisString
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *zlEntry
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "OneByteBinLength",
			args: args{
				str: &rs.RedisString{
					Content: make([]byte, OneByteBinLength),
				},
			},
			want: &zlEntry{
				encodingLength: 1,
				encodingVal:    []byte{0x3f},
				dataLength:     0x3f,
				data:           make([]byte, OneByteBinLength),
				encodingType:   binData,
			},
		},
		{
			name: "TwoByteBinLength",
			args: args{
				str: &rs.RedisString{
					Content: make([]byte, TwoByteBinLength),
				},
			},
			want: &zlEntry{
				encodingLength: 2,
				encodingVal:    []byte{0x7f, 0xff},
				dataLength:     TwoByteBinLength,
				data:           make([]byte, TwoByteBinLength),
				encodingType:   binData,
			},
		},
		{
			name: "FourByteBinLength",
			args: args{
				str: &rs.RedisString{
					Content: make([]byte, 0x4000),
				},
			},
			want: &zlEntry{
				encodingLength: 5,
				encodingVal:    []byte{0x80, 0x00, 0x00, 0x40, 0x00},
				dataLength:     0x4000,
				data:           make([]byte, 0x4000),
				encodingType:   binData,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			got, err := zl.transBinToEntry(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("ZipList.transBinToEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ZipList.transBinToEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_encodeEntryToBytes(t *testing.T) {
	type args struct {
		entry *zlEntry
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := encodeEntryToBytes(tt.args.entry); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encodeEntryToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_opPointer(t *testing.T) {
	type args struct {
		src    *byte
		offset uint32
		op     Op
	}
	tests := []struct {
		name    string
		args    args
		want    *byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := opPointer(tt.args.src, tt.args.offset, tt.args.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("opPointer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("opPointer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBigEndianUint24(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BigEndianUint24(tt.args.bytes); got != tt.want {
				t.Errorf("BigEndianUint24() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBigEndianPutUint24(t *testing.T) {
	type args struct {
		bytes []byte
		num   uint32
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "INT24_MIN",
			args: args{
				bytes: make([]byte, 3),
				// num: uint32(int32(utils.INT24_MIN)),
				num: uint32(0x800000),
			},
		},
		{
			name: "INT24_MAX",
			args: args{
				bytes: make([]byte, 3),
				num:   uint32(utils.INT24_MAX),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			BigEndianPutUint24(tt.args.bytes, tt.args.num)
			t.Logf("bytes: %s\n", hex.EncodeToString(tt.args.bytes))
		})
	}
}

func Test_zlEntry_getLength(t *testing.T) {
	type fields struct {
		offset          uint32
		prevFieldLength uint32
		prevVal         []byte
		encodingLength  uint32
		encodingVal     []byte
		dataLength      uint32
		data            []byte
		encodingType    dataType
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &zlEntry{
				offset:          tt.fields.offset,
				prevFieldLength: tt.fields.prevFieldLength,
				prevVal:         tt.fields.prevVal,
				encodingLength:  tt.fields.encodingLength,
				encodingVal:     tt.fields.encodingVal,
				dataLength:      tt.fields.dataLength,
				data:            tt.fields.data,
				encodingType:    tt.fields.encodingType,
			}
			if got := entry.getLength(); got != tt.want {
				t.Errorf("zlEntry.getLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_zlEntry_getPrevLength(t *testing.T) {
	type fields struct {
		offset          uint32
		prevFieldLength uint32
		prevVal         []byte
		encodingLength  uint32
		encodingVal     []byte
		dataLength      uint32
		data            []byte
		encodingType    dataType
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &zlEntry{
				offset:          tt.fields.offset,
				prevFieldLength: tt.fields.prevFieldLength,
				prevVal:         tt.fields.prevVal,
				encodingLength:  tt.fields.encodingLength,
				encodingVal:     tt.fields.encodingVal,
				dataLength:      tt.fields.dataLength,
				data:            tt.fields.data,
				encodingType:    tt.fields.encodingType,
			}
			if got := entry.getPrevLength(); got != tt.want {
				t.Errorf("zlEntry.getPrevLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZipList_getLength(t *testing.T) {
	type fields struct {
		content []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			if got := zl.getLength(); got != tt.want {
				t.Errorf("ZipList.getLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZipList_getOffset(t *testing.T) {
	type fields struct {
		content []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			if got := zl.getOffset(); got != tt.want {
				t.Errorf("ZipList.getOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZipList_ZipListInsert(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		position *byte
		val      interface{}
	}

	content := []byte{0x00, 0x00, 0x00, 0x11,
		0x00, 0x00, 0x00, 0x0c,
		0x00, 0x01,
		0x00, 0xf0,
		0x02, 0x02, 0xff, 0xff,
		0xff}

	resultContent := []byte{0x00, 0x00, 0x00, 0x14,
		0x00, 0x00, 0x00, 0x0f,
		0x00, 0x02,
		0x00, 0xf0,
		0x02, EncodingInt8, 0x0d,
		0x03, 0x02, 0xff, 0xff,
		0xff}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "case1",
			fields: fields{
				content: content,
			},
			args: args{
				position: &content[10],
				val:      int64(13),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			if err := zl.ZipListInsert(tt.args.position, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("ZipList.ZipListInsert() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(zl.content, resultContent) {
				t.Errorf("ZipList.ZipListInsert() got = %v, want = %v", zl.content, resultContent)
			}
		})
	}
}

func TestZipList_convertValToEntry(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		val interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *zlEntry
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			got, err := zl.convertValToEntry(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ZipList.convertValToEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ZipList.convertValToEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZipList_fillPrev(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		position *byte
		entry    *zlEntry
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			zl.fillPrev(tt.args.position, tt.args.entry)
		})
	}
}

func TestZipList_insertEntry(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		position *byte
		entry    *zlEntry
	}
	content := []byte{0x00, 0x00, 0x00, 0x13,
		0x00, 0x00, 0x00, 0x0f,
		0x00, 0x02,
		0x00, 0xf0,
		0x02, EncodingInt8, 0x0d,
		0x03, EncodingInt8, 0x0e,
		0xff}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "case1",
			fields: fields{
				content: content,
			},
			args: args{
				entry: &zlEntry{
					offset:          12,
					prevFieldLength: 1,
					prevVal:         []byte{0x02},
					encodingLength:  2,
					encodingVal:     []byte{0x40, 0xff},
					dataLength:      0xff,
					data:            make([]byte, 0xff),
				},
				position: &content[10],
			},
		},
	}

	resultContent := make([]byte, len(content)+int(tests[0].args.entry.getLength())+4)
	binary.BigEndian.PutUint32(resultContent[:4], uint32(len(resultContent)))
	binary.BigEndian.PutUint32(resultContent[4:8], uint32(277))
	binary.BigEndian.PutUint16(resultContent[8:10], uint16(3))
	resultContent[10] = 0x00
	resultContent[11] = 0xf0
	resultContent[12] = 0x02
	resultContent[13] = 0x40
	resultContent[14] = 0xff
	resultContent[270] = 0xfe
	binary.BigEndian.PutUint32(resultContent[271:275], uint32(258))
	resultContent[275] = EncodingInt8
	resultContent[276] = 0x0d
	resultContent[277] = 7
	resultContent[278] = EncodingInt8
	resultContent[279] = 0x0e
	resultContent[len(resultContent)-1] = Tail

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			zl.insertEntry(tt.args.position, tt.args.entry)
			if !reflect.DeepEqual(zl.content, resultContent) {
				t.Errorf("want: %v, got: %v\n", resultContent, zl.content)
			}
		})
	}
}

func TestZipList_fillContent(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		entry        *zlEntry
		nextPosition *byte
	}
	content := []byte{0x00, 0x00, 0x00, 0x13,
		0x00, 0x00, 0x00, 0x0f,
		0x00, 0x02,
		0x00, 0xf0,
		0x02, EncodingInt8, 0x0d,
		0x03, EncodingInt8, 0x0e,
		0xff}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "case1",
			fields: fields{
				content: content,
			},
			args: args{
				entry: &zlEntry{
					offset:          12,
					prevFieldLength: 1,
					prevVal:         []byte{0x02},
					encodingLength:  2,
					encodingVal:     []byte{0x40, 0xff},
					dataLength:      0xff,
					data:            make([]byte, 0xff),
				},
				nextPosition: &content[12],
			},
		},
	}

	resultContent := make([]byte, len(content)+int(tests[0].args.entry.getLength())+4)
	// fillContent 未调整首部
	binary.BigEndian.PutUint32(resultContent[:4], uint32(len(content)))
	binary.BigEndian.PutUint32(resultContent[4:8], uint32(15))
	binary.BigEndian.PutUint16(resultContent[8:10], uint16(2))
	resultContent[10] = 0x00
	resultContent[11] = 0xf0
	resultContent[12] = 0x02
	resultContent[13] = 0x40
	resultContent[14] = 0xff
	resultContent[270] = 0xfe
	binary.BigEndian.PutUint32(resultContent[271:275], uint32(258))
	resultContent[275] = EncodingInt8
	resultContent[276] = 0x0d
	resultContent[277] = 7
	resultContent[278] = EncodingInt8
	resultContent[279] = 0x0e
	resultContent[len(resultContent)-1] = Tail

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			zl.copyContentForInsert(tt.args.entry, tt.args.nextPosition)
			if !reflect.DeepEqual(zl.content, resultContent) {
				t.Errorf("want: %v, got: %v\n", resultContent, zl.content)
			}
		})
	}
}

func TestZipList_calcBytesToExtend(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		prevLen uint32
		p       *byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*zlEntry
		want1  uint32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			got, got1 := zl.calcBytesToExtend(tt.args.prevLen, tt.args.p)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ZipList.calcBytesToExtend() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ZipList.calcBytesToExtend() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestZipList_whetherToExtend(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		prevLen uint32
		p       *byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			if got := zl.whetherToExtend(tt.args.prevLen, tt.args.p); got != tt.want {
				t.Errorf("ZipList.whetherToExtend() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZipList_decodeFieldPrev(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		p     *byte
		entry *zlEntry
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			zl.decodeFieldPrev(tt.args.p, tt.args.entry)
		})
	}
}

func TestZipList_copyContentForInsert(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		entry        *zlEntry
		nextPosition *byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			if got := zl.copyContentForInsert(tt.args.entry, tt.args.nextPosition); got != tt.want {
				t.Errorf("ZipList.copyContentForInsert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZipList_ZipListDelete(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		val interface{}
	}
	resultContent := []byte{0x00, 0x00, 0x00, 0x14,
		0x00, 0x00, 0x00, 0x0f,
		0x00, 0x02,
		0x00, 0xf0,
		0x02, EncodingInt8, 0x0d,
		0x03, 0x02, 0xff, 0xff,
		0xff}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
		wantFields fields
	}{
		// TODO: Add test cases.
		{
			name: "case1",
			fields: fields{
				content: resultContent,
			},
			args: args{
				val: int64(12),
			},
			want: false,
			wantFields: fields{
				content: resultContent,
			},
		},
		{
			name: "case2",
			fields: fields{
				content: resultContent,
			},
			args: args{
				val: int64(13),
			},
			want: true,
			wantFields: fields{
				content: []byte{0x00, 0x00, 0x00, 0x11,
					0x00, 0x00, 0x00, 0x0c,
					0x00, 0x01,
					0x00, 0xf0,
					0x02, 0x02, 0xff, 0xff,
					0xff},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			got, err := zl.ZipListDelete(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ZipList.ZipListDelete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ZipList.ZipListDelete() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(zl.content, tt.wantFields.content) {
				t.Errorf("ZipList.ZipListDelete() got = %v, want = %v", zl.content, tt.wantFields.content)
			}
		})
	}
}

func TestZipList_findElem(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		val interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *zlEntry
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			got, err := zl.findElem(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ZipList.findElem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ZipList.findElem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZipList_deleteEntry(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		entry *zlEntry
	}

	content := []byte{0x00, 0x00, 0x00, 0x13,
		0x00, 0x00, 0x00, 0x0f,
		0x00, 0x02,
		0x00, 0xf0,
		0x02, EncodingInt8, 0x0d,
		0x03, EncodingInt8, 0x0e,
		0xff}

	newEntry := &zlEntry{
		offset:          12,
		prevFieldLength: 1,
		prevVal:         []byte{0x02},
		encodingLength:  2,
		encodingVal:     []byte{0x40, 0xff},
		dataLength:      0xff,
		data:            make([]byte, 0xff),
	}

	resultContent := make([]byte, len(content)+int(newEntry.getLength())+4)
	binary.BigEndian.PutUint32(resultContent[:4], uint32(len(resultContent)))
	binary.BigEndian.PutUint32(resultContent[4:8], uint32(277))
	binary.BigEndian.PutUint16(resultContent[8:10], uint16(3))
	resultContent[10] = 0x00
	resultContent[11] = 0xf0
	resultContent[12] = 0x02
	resultContent[13] = 0x40
	resultContent[14] = 0xff
	resultContent[270] = 0xfe
	binary.BigEndian.PutUint32(resultContent[271:275], uint32(258))
	resultContent[275] = EncodingInt8
	resultContent[276] = 0x0d
	resultContent[277] = 7
	resultContent[278] = EncodingInt8
	resultContent[279] = 0x0e
	resultContent[len(resultContent)-1] = Tail

	offset := uint32(12)
	entry := &zlEntry{
		offset:          12,
		prevFieldLength: 1,
		prevVal:         resultContent[offset:offset+1],
		encodingLength:  2,
		encodingVal:     resultContent[offset+1: offset+3],
		dataLength:      0xff,
		data:            resultContent[offset+3: offset + 258],
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		wantFields fields
	}{
		// TODO: Add test cases.
		{
			name: "case1",
			fields: fields{
				content: resultContent,
			},
			args: args{
				entry: entry,
			},
			wantFields: fields{
				content: content,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			zl.deleteEntry(tt.args.entry)
			if !reflect.DeepEqual(zl.content, tt.wantFields.content) {
				t.Errorf("ZipList.deleteEntry() got = %v, want = %v", zl.content, tt.wantFields.content)
			}
		})
	}
}

func TestZipList_calcBytesToShrink(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		prevLen uint32
		p       *byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*zlEntry
		want1  uint32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			got, got1 := zl.calcBytesToShrink(tt.args.prevLen, tt.args.p)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ZipList.calcBytesToShrink() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ZipList.calcBytesToShrink() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestZipList_whetherToShrink(t *testing.T) {
	type fields struct {
		content []byte
	}
	type args struct {
		prevLen uint32
		p       *byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZipList{
				content: tt.fields.content,
			}
			if got := zl.whetherToShrink(tt.args.prevLen, tt.args.p); got != tt.want {
				t.Errorf("ZipList.whetherToShrink() = %v, want %v", got, tt.want)
			}
		})
	}
}
