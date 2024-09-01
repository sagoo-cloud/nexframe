package convert

import (
	"math"
	"testing"
)

// 用于测试的自定义类型
type customUint64 int

func (c customUint64) Uint64() uint64 {
	return uint64(c)
}

func TestUintConversions(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantUint uint
		wantU8   uint8
		wantU16  uint16
		wantU32  uint32
		wantU64  uint64
	}{
		{"nil", nil, 0, 0, 0, 0, 0},
		{"uint", uint(123), 123, 123, 123, 123, 123},
		{"uint8", uint8(255), 255, 255, 255, 255, 255},
		{"uint16", uint16(65535), 65535, 255, 65535, 65535, 65535},
		{"uint32", uint32(4294967295), 4294967295, 255, 65535, 4294967295, 4294967295},
		{"uint64", uint64(18446744073709551615), 18446744073709551615, 255, 65535, 4294967295, 18446744073709551615},
		{"int", int(-1), math.MaxUint64, 255, 65535, 4294967295, 18446744073709551615},
		{"int8", int8(-1), 255, 255, 255, 255, 255},
		{"int16", int16(-1), 65535, 255, 65535, 65535, 65535},
		{"int32", int32(-1), 4294967295, 255, 65535, 4294967295, 4294967295},
		{"int64", int64(-1), 18446744073709551615, 255, 65535, 4294967295, 18446744073709551615},
		{"float32", float32(123.45), 123, 123, 123, 123, 123},
		{"float64", 123.45, 123, 123, 123, 123, 123},
		{"bool true", true, 1, 1, 1, 1, 1},
		{"bool false", false, 0, 0, 0, 0, 0},
		{"string decimal", "12345", 12345, 255, 12345, 12345, 12345},
		{"string hex", "0xFF", 255, 255, 255, 255, 255},
		{"string float", "123.45", 123, 123, 123, 123, 123},
		{"[]byte", []byte{0xFF, 0xFF, 0xFF, 0xFF}, 4294967295, 255, 65535, 4294967295, 4294967295},
		{"customUint64", customUint64(12345), 12345, 255, 12345, 12345, 12345},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Uint(tt.input); got != tt.wantUint {
				t.Errorf("Uint() = %v, want %v", got, tt.wantUint)
			}
			if got := Uint8(tt.input); got != tt.wantU8 {
				t.Errorf("Uint8() = %v, want %v", got, tt.wantU8)
			}
			if got := Uint16(tt.input); got != tt.wantU16 {
				t.Errorf("Uint16() = %v, want %v", got, tt.wantU16)
			}
			if got := Uint32(tt.input); got != tt.wantU32 {
				t.Errorf("Uint32() = %v, want %v", got, tt.wantU32)
			}
			if got := Uint64(tt.input); got != tt.wantU64 {
				t.Errorf("Uint64() = %v, want %v", got, tt.wantU64)
			}
		})
	}
}

// 性能测试
func BenchmarkUintConversions(b *testing.B) {
	inputs := []interface{}{
		uint(123),
		int(-1),
		float64(123.45),
		"12345",
		"0xFF",
		[]byte{0xFF, 0xFF, 0xFF, 0xFF},
		true,
		customUint64(12345),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			Uint(input)
			Uint8(input)
			Uint16(input)
			Uint32(input)
			Uint64(input)
		}
	}
}
