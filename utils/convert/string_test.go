package convert

import (
	"testing"
	"time"
)

// 用于测试的自定义类型
type customString string

func (c customString) String() string {
	return string(c) + "_custom"
}

type customError struct{}

func (c customError) Error() string {
	return "custom_error"
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil", nil, ""},
		{"empty string", "", ""},
		{"string", "hello", "hello"},
		{"int", 123, "123"},
		{"int8", int8(8), "8"},
		{"int16", int16(16), "16"},
		{"int32", int32(32), "32"},
		{"int64", int64(64), "64"},
		{"uint", uint(123), "123"},
		{"uint8", uint8(8), "8"},
		{"uint16", uint16(16), "16"},
		{"uint32", uint32(32), "32"},
		{"uint64", uint64(64), "64"},
		{"float32", float32(3.14), "3.14"},
		{"float64", 3.14159, "3.14159"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"[]byte", []byte("bytes"), "bytes"},
		{"time.Time", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), "2023-01-01T00:00:00Z"},
		{"*int nil", (*int)(nil), ""},
		{"customString", customString("test"), "test_custom"},
		{"customError", customError{}, "custom_error"},
		{"slice", []int{1, 2, 3}, "[1,2,3]"},
		{"map", map[string]int{"a": 1, "b": 2}, "{\"a\":1,\"b\":2}"},
		{"struct", struct{ Name string }{"John"}, "{\"Name\":\"John\"}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := String(tt.input)
			if result != tt.expected {
				t.Errorf("String(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// 性能测试
func BenchmarkString(b *testing.B) {
	inputs := []interface{}{
		123,
		"hello",
		3.14,
		true,
		[]int{1, 2, 3},
		map[string]int{"a": 1, "b": 2},
		time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			String(input)
		}
	}
}
