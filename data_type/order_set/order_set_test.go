package orderset

import (
	"reflect"
	"testing"

	"github.com/WANGgbin/tiny_redis/data_type/redis_string"
)

func TestCreatSkipList(t *testing.T) {
	type args struct {
		scoreValPairs []*ScoreValPair
	}
	tests := []struct {
		name string
		args args
		want []*ScoreValPair
	}{
		{
			name: "case1",
			args: args{
				scoreValPairs: []*ScoreValPair{
					{
						Score: 1,
						Val: redis_string.RedisString{
							Content: []byte("12"),
						},
					},
					{
						Score: 1,
						Val: redis_string.RedisString{
							Content: []byte("1"),
						},
					},
					{
						Score: 4,
						Val: redis_string.RedisString{
							Content: []byte("4"),
						},
					},
					{
						Score: 2,
						Val: redis_string.RedisString{
							Content: []byte("2"),
						},
					},
					{
						Score: -1.0,
						Val: redis_string.RedisString{
							Content: []byte("-1.0"),
						},
					},
				},
			},
			want: []*ScoreValPair{
				{
					Score: -1.0,
					Val: redis_string.RedisString{
						Content: []byte("-1.0"),
					},
				},
				{
					Score: 1,
					Val: redis_string.RedisString{
						Content: []byte("1"),
					},
				},
				{
					Score: 1,
					Val: redis_string.RedisString{
						Content: []byte("12"),
					},
				},
				{
					Score: 2,
					Val: redis_string.RedisString{
						Content: []byte("2"),
					},
				},
				{
					Score: 4,
					Val: redis_string.RedisString{
						Content: []byte("4"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreatSkipList(tt.args.scoreValPairs)
			scoreValPairs := got.GetAllScoreValPairs()
			for index, scoreVal := range scoreValPairs {
				equal, _ := redis_string.StrCmp(&scoreVal.Val, &tt.want[index].Val)
				if equal != int8(0) || scoreVal.Score != tt.want[index].Score {
					t.Fatalf("index: %d, want: %v, got: %v\n", index, tt.want[index], scoreVal)
				}
			}
		})
	}
}

func TestInitSkipList(t *testing.T) {
	tests := []struct {
		name string
		want *SkipList
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitSkipList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitSkipList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkipList_InsertNode(t *testing.T) {
	type fields struct {
		header *SkipListNode
		tail   *SkipListNode
		level  int8
		length int64
	}
	type args struct {
		scoreValPair *ScoreValPair
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *SkipListNode
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &SkipList{
				header: tt.fields.header,
				tail:   tt.fields.tail,
				level:  tt.fields.level,
				length: tt.fields.length,
			}
			got, err := sp.InsertNode(tt.args.scoreValPair)
			if (err != nil) != tt.wantErr {
				t.Errorf("SkipList.InsertNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SkipList.InsertNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkipList_getLastLessNode(t *testing.T) {
	type fields struct {
		header *SkipListNode
		tail   *SkipListNode
		level  int8
		length int64
	}
	type args struct {
		scoreValPair *ScoreValPair
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*SkipListNode
		want1   []int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &SkipList{
				header: tt.fields.header,
				tail:   tt.fields.tail,
				level:  tt.fields.level,
				length: tt.fields.length,
			}
			got, got1, err := sp.getLastLessNode(tt.args.scoreValPair)
			if (err != nil) != tt.wantErr {
				t.Errorf("SkipList.getLastLessNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SkipList.getLastLessNode() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("SkipList.getLastLessNode() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getRandomLevel(t *testing.T) {
	tests := []struct {
		name string
		want int8
	}{
		{
			name: "case1",
		},
		{
			name: "case2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getRandomLevel()
			if got > SkipListMaxLevel {
				t.Errorf("getRandomLevel() = %d, should less then %d", got, SkipListMaxLevel)
			} else {
				t.Logf("getRandomLevel() = %d", got)
			}
		})
	}
}

func Test_createNewNode(t *testing.T) {
	type args struct {
		level        int8
		scoreValPair *ScoreValPair
	}
	tests := []struct {
		name string
		args args
		want *SkipListNode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createNewNode(tt.args.level, tt.args.scoreValPair); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createNewNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkipList_DeleteNode(t *testing.T) {
	// 构造一个 sp
	scoreValPairs := []*ScoreValPair{
		{
			Score: 1,
			Val: redis_string.RedisString{
				Content: []byte("12"),
			},
		},
		{
			Score: 1,
			Val: redis_string.RedisString{
				Content: []byte("1"),
			},
		},
		{
			Score: 4,
			Val: redis_string.RedisString{
				Content: []byte("4"),
			},
		},
		{
			Score: 2,
			Val: redis_string.RedisString{
				Content: []byte("2"),
			},
		},
		{
			Score: -1.0,
			Val: redis_string.RedisString{
				Content: []byte("-1.0"),
			},
		},
	}

	sp := CreatSkipList(scoreValPairs)

	for _, scoreValPair := range scoreValPairs {
		result := sp.DeleteNode(scoreValPair)
		if !result {
			t.Fatalf("delete scoreValPair: %v failed, current sp: %v", scoreValPair, sp.GetAllScoreValPairs())
		}
		t.Logf("After delete scoreValPair: %v, current sp: %v and pairs:%v", scoreValPair, sp, sp.GetAllScoreValPairs())
	}

	// sp 为空
	result := sp.DeleteNode(scoreValPairs[0])
	if result {
		t.Fatalf("When sp has no nodes, but delete scoreValPair: %v success", scoreValPairs[0])
	}
}

func TestSkipList_UpdateNode(t *testing.T) {
	// 构造一个 sp
	scoreValPairs := []*ScoreValPair{
		{
			Score: 1,
			Val: redis_string.RedisString{
				Content: []byte("12"),
			},
		},
		{
			Score: 1,
			Val: redis_string.RedisString{
				Content: []byte("1"),
			},
		},
		{
			Score: 4,
			Val: redis_string.RedisString{
				Content: []byte("4"),
			},
		},
		{
			Score: 2,
			Val: redis_string.RedisString{
				Content: []byte("2"),
			},
		},
		{
			Score: -1.0,
			Val: redis_string.RedisString{
				Content: []byte("-1.0"),
			},
		},
	}

	sp := CreatSkipList(scoreValPairs)

	for _, scoreValPair := range scoreValPairs {
		_, err := sp.UpdateNode(scoreValPair, 0.0)
		if err != nil {
			printAllScoreValPairs(sp, t)
			t.Fatalf("update scoreValPair: %v failed", scoreValPair)
		}
		t.Logf("After updating scoreValPair: %v", scoreValPair)
		printAllScoreValPairs(sp, t)
	}
}

func printAllScoreValPairs(sp *SkipList, t *testing.T) {
	pairs := sp.GetAllScoreValPairs()
	t.Log("Current content:\n")
	for index, pair := range pairs {
		t.Logf("	index:[%d], pair: {Score: %.f, Val: %s}\n", index, pair.Score, string(pair.Val.Content))
	}
}
