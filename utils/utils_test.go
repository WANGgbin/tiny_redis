package utils

import "testing"

func TestConst(t *testing.T) {
	t.Logf("INT8_MAX: %d, INT8_MIN: %d", INT8_MAX, INT8_MIN)
	t.Logf("INT16_MAX: %d, INT16_MIN: %d", INT16_MAX, INT16_MIN)
	t.Logf("INT24_MAX: %d, INT24_MIN: %d", INT24_MAX, INT24_MIN)
	t.Logf("INT32_MAX: %d, INT32_MIN: %d", INT32_MAX, INT32_MIN)
	t.Logf("INT64_MAX: %d, INT64_MIN: %d", INT64_MAX, INT64_MIN)
}
