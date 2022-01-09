package orderset

import (
	"log"
	"math/rand"
	"time"

	rs "github.com/WANGgbin/tiny_redis/data_type/redis_string"
)

// skipList 实现 orderset

const (
	SkipListMaxLevel int8 = 32
)

type SkipList struct {
	header *SkipListNode
	tail   *SkipListNode
	level  int8
	length int64
}

type SkipListNode struct {
	indexes      []*IndexNode
	score        float64
	val          rs.RedisString
	backwardNode *SkipListNode
}

type IndexNode struct {
	forwardNode *SkipListNode
	span        int64
}

type ScoreValPair struct {
	Score float64
	Val   rs.RedisString
}

func cmp(src, dst *ScoreValPair) int8 {
	if src.Score != dst.Score {
		if src.Score > dst.Score {
			return int8(1)
		}
		return int8(-1)
	}
	result, _ := rs.StrCmp(&src.Val, &dst.Val)
	return result
}

func CreatSkipList(scoreValPairs []*ScoreValPair) *SkipList {
	sp := InitSkipList()
	for _, pair := range scoreValPairs {
		sp.InsertNode(pair)
	}

	return sp
}

func InitSkipList() *SkipList {
	header := &SkipListNode{
		indexes: make([]*IndexNode, SkipListMaxLevel),
	}

	for i := range header.indexes {
		header.indexes[i] = new(IndexNode)
	}

	sp := &SkipList{
		header: header,
		tail:   header,
		level:  int8(1),
	}

	return sp
}

// GetAllScoreValPairs get all elements in sp
func (sp *SkipList) GetAllScoreValPairs() []*ScoreValPair {
	if sp.length == 0 {
		return nil
	}
	ret := make([]*ScoreValPair, sp.length)
	curNode := sp.header.indexes[0].forwardNode

	for i := 0; i < int(sp.length); i++ {
		ret[i] = &ScoreValPair{
			Score: curNode.score,
			Val: rs.RedisString{
				Content: make([]byte, len(curNode.val.Content)),
			},
		}
		copy(ret[i].Val.Content, curNode.val.Content)
		curNode = curNode.indexes[0].forwardNode
	}

	return ret
}

// InsertNode inserts a new node in sp
func (sp *SkipList) InsertNode(scoreValPair *ScoreValPair) (*SkipListNode, error) {
	lastLessNodes, rank, err := sp.getLastLessNode(scoreValPair)
	if err != nil {
		return nil, err
	}

	newLevel := getRandomLevel()
	log.Printf("newLevel: %d, scoreValPair: %v", newLevel, scoreValPair)
	if newLevel > sp.level {
		for curLevel := sp.level; curLevel < newLevel; curLevel++ {
			lastLessNodes[curLevel] = sp.header
			rank[curLevel] = 0
		}
		sp.level = newLevel
	}

	newNode := createNewNode(newLevel, scoreValPair)

	// 调整每一曾结点的 forward 和 span
	for curLevel := 0; curLevel < int(newLevel); curLevel++ {
		lessNode := lastLessNodes[curLevel]
		nextNode := lessNode.indexes[curLevel].forwardNode
		span := lessNode.indexes[curLevel].span
		newNode.indexes[curLevel].forwardNode = nextNode

		if nextNode == nil {
			newNode.indexes[curLevel].span = 0
		} else {
			newNode.indexes[curLevel].span = span - (rank[0] - rank[curLevel])
		}

		lessNode.indexes[curLevel].forwardNode = newNode
		lessNode.indexes[curLevel].span = rank[0] - rank[curLevel] + 1
	}

	// 调整 > newLevel 层
	for curLevel := newLevel; curLevel < sp.level; curLevel++ {
		if lastLessNodes[curLevel].indexes[curLevel].forwardNode != nil {
			lastLessNodes[curLevel].indexes[curLevel].span++
		}
	}

	// 插入结点是最后一个结点
	if newNode.indexes[0].forwardNode == nil {
		sp.tail = newNode
	} else {
		newNode.indexes[0].forwardNode.backwardNode = newNode
	}

	// 调整插入节点 backward
	if lastLessNodes[0] == sp.header {
		newNode.backwardNode = nil
	} else {
		newNode.backwardNode = lastLessNodes[0]
	}

	sp.length++

	return newNode, nil
}

func (sp *SkipList) getLastLessNode(scoreValPair *ScoreValPair) ([]*SkipListNode, []int64, error) {
	// 存放最后一个小于待插入节点的节点
	lastLessNode := make([]*SkipListNode, SkipListMaxLevel)
	// 存放 lastLessNode对应节点的 rank
	rank := make([]int64, SkipListMaxLevel)

	curNode := sp.header
	curRank := int64(0)

	for curLevel := sp.level - 1; curLevel >= 0; curLevel-- {
		index := curNode.indexes[curLevel]
		for {
			nextNode := index.forwardNode
			if nextNode == nil {
				lastLessNode[curLevel] = curNode
				rank[curLevel] = curRank
				break
			}

			cmp, err := rs.StrCmp(&nextNode.val, &scoreValPair.Val)
			if err != nil {
				return nil, nil, err
			}
			if nextNode.score < scoreValPair.Score || (nextNode.score == scoreValPair.Score && cmp < int8(0)) {
				curRank += index.span
				curNode = nextNode
				index = curNode.indexes[curLevel]
			} else {
				lastLessNode[curLevel] = curNode
				rank[curLevel] = curRank
				break
			}
		}
	}

	return lastLessNode, rank, nil
}

func getRandomLevel() int8 {
	level := int8(1)
	rand.Seed(time.Now().UnixNano())

	for {
		if rand.Intn(2) == 0 {
			break
		}

		if level >= SkipListMaxLevel {
			break
		}
		level += 1
	}

	return level
}

func createNewNode(level int8, scoreValPair *ScoreValPair) *SkipListNode {
	newNode := &SkipListNode{
		score: scoreValPair.Score,
		val: rs.RedisString{
			Content: make([]byte, len(scoreValPair.Val.Content)),
		},
		indexes: make([]*IndexNode, level),
	}

	copy(newNode.val.Content, scoreValPair.Val.Content)

	for i := 0; i < int(level); i++ {
		newNode.indexes[i] = new(IndexNode)
	}

	return newNode
}

// DeleteNode deletes node with specific score and val
func (sp *SkipList) DeleteNode(scoreValPair *ScoreValPair) bool {
	// 判断是否存在
	lastLessNodes, _, _ := sp.getLastLessNode(scoreValPair)
	nextNode := lastLessNodes[0].indexes[0].forwardNode

	if nextNode == nil {
		log.Printf("There is no %v in sp", scoreValPair)
		return false
	}

	cmp, _ := rs.StrCmp(&nextNode.val, &scoreValPair.Val)
	if nextNode.score != scoreValPair.Score || cmp != 0 {
		return false
	}

	// 存在则删除
	for curLevel := 0; curLevel < len(nextNode.indexes); curLevel++ {
		lastLessNodes[curLevel].indexes[curLevel].forwardNode = nextNode.indexes[curLevel].forwardNode
		lastLessNodes[curLevel].indexes[curLevel].span += nextNode.indexes[curLevel].span - 1

	}

	for curLevel := len(nextNode.indexes); curLevel < int(sp.level); curLevel++ {
		if lastLessNodes[curLevel].indexes[curLevel].forwardNode != nil {
			lastLessNodes[curLevel].indexes[curLevel].span -= 1
		}
	}

	// 调整跳表 level
	for curLevel := int(sp.level); curLevel >= 0; curLevel-- {
		if lastLessNodes[curLevel] == sp.header && lastLessNodes[curLevel].indexes[curLevel].forwardNode == nil {
			sp.level--
		}
	}

	// 调整 backward
	if nextNode.indexes[0].forwardNode == nil {
		sp.tail = lastLessNodes[0]
	} else {
		if lastLessNodes[0] == sp.header {
			nextNode.indexes[0].forwardNode.backwardNode = nil
		} else {
			nextNode.indexes[0].forwardNode.backwardNode = lastLessNodes[0]
		}
	}

	sp.length--
	// 特别注意，当不包含元素的时候，level 重新调整为 1 而不是 0
	if sp.length == 0 {
		sp.level = 1
	}

	return true
}

// UpdateNode updates node with specide score and val
func (sp *SkipList) UpdateNode(scoreValPair *ScoreValPair, newScore float64) (*SkipListNode, error) {
	targetNode := sp.isExist(scoreValPair)
	newScoreValPair := &ScoreValPair{Score: newScore, Val: scoreValPair.Val}

	if targetNode == nil {
		return sp.InsertNode(newScoreValPair)
	}

	backNode := targetNode.backwardNode
	forwardNode := targetNode.indexes[0].forwardNode
	if (backNode == nil || cmp(&ScoreValPair{Score: backNode.score, Val: backNode.val}, newScoreValPair) == -1) &&
		(forwardNode == nil || cmp(&ScoreValPair{Score: forwardNode.score, Val: forwardNode.val}, newScoreValPair) == 1) {
		targetNode.score = newScore
		return targetNode, nil
	}

	sp.DeleteNode(scoreValPair)

	return sp.InsertNode(newScoreValPair)
}

// isExist is used to judge whether scorevalPair exists
func (sp *SkipList) isExist(scoreValPair *ScoreValPair) *SkipListNode {
	lastLessNodes, _, _ := sp.getLastLessNode(scoreValPair)
	nextNode := lastLessNodes[0].indexes[0].forwardNode

	if nextNode == nil {
		log.Printf("There is no %v in sp", scoreValPair)
		return nil
	}

	cmp, _ := rs.StrCmp(&nextNode.val, &scoreValPair.Val)
	if nextNode.score != scoreValPair.Score || cmp != 0 {
		return nil
	}

	return nextNode
}

func cmpSkipListNode(src, dst *SkipListNode) int8 {
	return cmp(&ScoreValPair{Score: src.score, Val: src.val}, &ScoreValPair{Score: dst.score, Val: dst.val})
}

// TODO: implement other useful apis about skiplist
