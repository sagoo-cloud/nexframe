// Package gbinary 提供了二进制数据处理功能。
package gbinary

// NOTE: THIS IS AN EXPERIMENTAL FEATURE!

// Bit 表示二进制位(0或1)。
type Bit int8

// EncodeBits 将整数编码为指定长度的二进制位数组。
// 本质上是调用EncodeBitsWithUint，将int转换为uint进行处理。
func EncodeBits(bits []Bit, i int, l int) []Bit {
	return EncodeBitsWithUint(bits, uint(i), l)
}

// EncodeBitsWithUint 将无符号整数按位编码到二进制位数组中。
// bits: 现有的位数组，如果为nil则创建新数组
// ui: 要编码的无符号整数
// l: 编码后占用的位数
// 返回包含编码结果的位数组
func EncodeBitsWithUint(bits []Bit, ui uint, l int) []Bit {
	a := make([]Bit, l)
	for i := l - 1; i >= 0; i-- {
		a[i] = Bit(ui & 1)
		ui >>= 1
	}
	if bits != nil {
		return append(bits, a...)
	}
	return a
}

// EncodeBitsToBytes 将二进制位数组编码为字节切片。
// 从左到右编码，如果位数不足8的倍数，末尾补0。
func EncodeBitsToBytes(bits []Bit) []byte {
	if len(bits)%8 != 0 {
		for i := 0; i < len(bits)%8; i++ {
			bits = append(bits, 0)
		}
	}
	b := make([]byte, 0)
	for i := 0; i < len(bits); i += 8 {
		b = append(b, byte(DecodeBitsToUint(bits[i:i+8])))
	}
	return b
}

// DecodeBits 将二进制位数组解码为整数。
func DecodeBits(bits []Bit) int {
	v := 0
	for _, i := range bits {
		v = v<<1 | int(i)
	}
	return v
}

// DecodeBitsToUint 将二进制位数组解码为无符号整数。
func DecodeBitsToUint(bits []Bit) uint {
	v := uint(0)
	for _, i := range bits {
		v = v<<1 | uint(i)
	}
	return v
}

// DecodeBytesToBits 将字节切片解析为二进制位数组。
// 每个字节被解析为8个二进制位。
func DecodeBytesToBits(bs []byte) []Bit {
	bits := make([]Bit, 0)
	for _, b := range bs {
		bits = EncodeBitsWithUint(bits, uint(b), 8)
	}
	return bits
}
