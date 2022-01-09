package redis_string

import (
	"errors"
	"fmt"
)

// string 实现类似与 C++ string

type RedisString struct {
	Content []byte
}

func StrCmp(src, dst *RedisString) (int8, error) {
	var errInfo string
	if src == nil || dst == nil {
		errInfo = fmt.Sprintf("to compare string, src and dst must't be nil")
		return int8(-1), errors.New(errInfo)
	}

	toCmpLen := len(src.Content)
	if len(src.Content) > len(dst.Content) {
		toCmpLen = len(dst.Content)
	}

	// FIXME: 小写字母 ascii 大于 大写字母 ascii
	for index := 0; index < toCmpLen; index++ {
		if src.Content[index] > dst.Content[index] {
			return int8(1), nil
		}

		if src.Content[index] < dst.Content[index] {
			return int8(-1), nil
		}
	}

	if len(src.Content) > len(dst.Content) {
		return int8(1), nil
	}

	if len(src.Content) < len(dst.Content) {
		return int8(-1), nil
	}

	return int8(0), nil

}
