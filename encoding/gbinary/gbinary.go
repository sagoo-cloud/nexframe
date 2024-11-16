// Package gbinary 提供了基本类型和字节切片之间的编码和解码功能。
package gbinary

// Encode 将任意数量的值编码为字节切片。
// 使用小端序进行编码。
func Encode(values ...interface{}) []byte {
	return LeEncode(values...)
}

// EncodeByLength 将指定数量的值编码为字节切片。
// 返回的字节切片长度由参数length指定。使用小端序编码。
func EncodeByLength(length int, values ...interface{}) []byte {
	return LeEncodeByLength(length, values...)
}

// Decode 将字节切片解码到指定的值中。
// 使用小端序进行解码。如果解码过程中发生错误，将返回相应的错误。
func Decode(b []byte, values ...interface{}) error {
	return LeDecode(b, values...)
}

// EncodeString 将字符串编码为字节切片。
// 使用小端序编码。
func EncodeString(s string) []byte {
	return LeEncodeString(s)
}

// DecodeToString 将字节切片解码为字符串。
// 使用小端序解码。
func DecodeToString(b []byte) string {
	return LeDecodeToString(b)
}

// EncodeBool 将布尔值编码为字节切片。
// 使用小端序编码。
func EncodeBool(b bool) []byte {
	return LeEncodeBool(b)
}

// EncodeInt 将int类型的整数编码为字节切片。
// 使用小端序编码。
func EncodeInt(i int) []byte {
	return LeEncodeInt(i)
}

// EncodeUint 将uint类型的无符号整数编码为字节切片。
// 使用小端序编码。
func EncodeUint(i uint) []byte {
	return LeEncodeUint(i)
}

// EncodeInt8 将int8类型的整数编码为字节切片。
// 使用小端序编码。
func EncodeInt8(i int8) []byte {
	return LeEncodeInt8(i)
}

// EncodeUint8 将uint8类型的无符号整数编码为字节切片。
// 使用小端序编码。
func EncodeUint8(i uint8) []byte {
	return LeEncodeUint8(i)
}

// EncodeInt16 将int16类型的整数编码为字节切片。
// 使用小端序编码。
func EncodeInt16(i int16) []byte {
	return LeEncodeInt16(i)
}

// EncodeUint16 将uint16类型的无符号整数编码为字节切片。
// 使用小端序编码。
func EncodeUint16(i uint16) []byte {
	return LeEncodeUint16(i)
}

// EncodeInt32 将int32类型的整数编码为字节切片。
// 使用小端序编码。
func EncodeInt32(i int32) []byte {
	return LeEncodeInt32(i)
}

// EncodeUint32 将uint32类型的无符号整数编码为字节切片。
// 使用小端序编码。
func EncodeUint32(i uint32) []byte {
	return LeEncodeUint32(i)
}

// EncodeInt64 将int64类型的整数编码为字节切片。
// 使用小端序编码。
func EncodeInt64(i int64) []byte {
	return LeEncodeInt64(i)
}

// EncodeUint64 将uint64类型的无符号整数编码为字节切片。
// 使用小端序编码。
func EncodeUint64(i uint64) []byte {
	return LeEncodeUint64(i)
}

// EncodeFloat32 将float32类型的浮点数编码为字节切片。
// 使用小端序编码。
func EncodeFloat32(f float32) []byte {
	return LeEncodeFloat32(f)
}

// EncodeFloat64 将float64类型的浮点数编码为字节切片。
// 使用小端序编码。
func EncodeFloat64(f float64) []byte {
	return LeEncodeFloat64(f)
}

// DecodeToInt 将字节切片解码为int类型的整数。
// 使用小端序解码。
func DecodeToInt(b []byte) int {
	return LeDecodeToInt(b)
}

// DecodeToUint 将字节切片解码为uint类型的无符号整数。
// 使用小端序解码。
func DecodeToUint(b []byte) uint {
	return LeDecodeToUint(b)
}

// DecodeToBool 将字节切片解码为布尔值。
// 使用小端序解码。
func DecodeToBool(b []byte) bool {
	return LeDecodeToBool(b)
}

// DecodeToInt8 将字节切片解码为int8类型的整数。
// 使用小端序解码。
func DecodeToInt8(b []byte) int8 {
	return LeDecodeToInt8(b)
}

// DecodeToUint8 将字节切片解码为uint8类型的无符号整数。
// 使用小端序解码。
func DecodeToUint8(b []byte) uint8 {
	return LeDecodeToUint8(b)
}

// DecodeToInt16 将字节切片解码为int16类型的整数。
// 使用小端序解码。
func DecodeToInt16(b []byte) int16 {
	return LeDecodeToInt16(b)
}

// DecodeToUint16 将字节切片解码为uint16类型的无符号整数。
// 使用小端序解码。
func DecodeToUint16(b []byte) uint16 {
	return LeDecodeToUint16(b)
}

// DecodeToInt32 将字节切片解码为int32类型的整数。
// 使用小端序解码。
func DecodeToInt32(b []byte) int32 {
	return LeDecodeToInt32(b)
}

// DecodeToUint32 将字节切片解码为uint32类型的无符号整数。
// 使用小端序解码。
func DecodeToUint32(b []byte) uint32 {
	return LeDecodeToUint32(b)
}

// DecodeToInt64 将字节切片解码为int64类型的整数。
// 使用小端序解码。
func DecodeToInt64(b []byte) int64 {
	return LeDecodeToInt64(b)
}

// DecodeToUint64 将字节切片解码为uint64类型的无符号整数。
// 使用小端序解码。
func DecodeToUint64(b []byte) uint64 {
	return LeDecodeToUint64(b)
}

// DecodeToFloat32 将字节切片解码为float32类型的浮点数。
// 使用小端序解码。
func DecodeToFloat32(b []byte) float32 {
	return LeDecodeToFloat32(b)
}

// DecodeToFloat64 将字节切片解码为float64类型的浮点数。
// 使用小端序解码。
func DecodeToFloat64(b []byte) float64 {
	return LeDecodeToFloat64(b)
}
