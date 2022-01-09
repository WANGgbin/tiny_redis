package utils

const (
	UINT8_MAX  = ^(uint8(0))
	UINT16_MAX = ^(uint16(0))
	UINT32_MAX = ^(uint32(0))
	UINT64_MAX = ^(uint64(0))

	INT8_MAX  = int8(^(uint8(1) << 7))
	INT8_MIN  = ^INT8_MAX
	INT16_MAX = int16(^(uint16(1) << 15))
	INT16_MIN = ^INT16_MAX
	INT24_MAX = int32(0x7fffff)
	INT24_MIN = -INT24_MAX - 1
	INT32_MAX = int32(^(uint32(1) << 31))
	INT32_MIN = ^INT32_MAX
	INT64_MAX = int64(^(uint64(1) << 63))
	INT64_MIN = ^INT64_MAX
)
