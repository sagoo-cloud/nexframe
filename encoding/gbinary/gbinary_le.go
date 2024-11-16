// Package gbinary 提供了基本类型和字节切片之间的编码和解码功能。
package gbinary

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/sagoo-cloud/nexframe/os/zlog/intlog"
	"github.com/sagoo-cloud/nexframe/utils/errors/gerror"
	"math"
)

// LeEncode 使用小端序将一个或多个值编码为字节切片。
// 通过类型断言检查values中每个值的类型，并调用相应的转换函数进行字节转换。
// 支持常见变量类型的断言，对于不支持的类型，使用fmt.Sprintf将值转换为字符串后再转换为字节。
func LeEncode(values ...interface{}) []byte {
	buf := new(bytes.Buffer)
	for i := 0; i < len(values); i++ {
		if values[i] == nil {
			return buf.Bytes()
		}
		switch value := values[i].(type) {
		case int:
			buf.Write(LeEncodeInt(value))
		case int8:
			buf.Write(LeEncodeInt8(value))
		case int16:
			buf.Write(LeEncodeInt16(value))
		case int32:
			buf.Write(LeEncodeInt32(value))
		case int64:
			buf.Write(LeEncodeInt64(value))
		case uint:
			buf.Write(LeEncodeUint(value))
		case uint8:
			buf.Write(LeEncodeUint8(value))
		case uint16:
			buf.Write(LeEncodeUint16(value))
		case uint32:
			buf.Write(LeEncodeUint32(value))
		case uint64:
			buf.Write(LeEncodeUint64(value))
		case bool:
			buf.Write(LeEncodeBool(value))
		case string:
			buf.Write(LeEncodeString(value))
		case []byte:
			buf.Write(value)
		case float32:
			buf.Write(LeEncodeFloat32(value))
		case float64:
			buf.Write(LeEncodeFloat64(value))

		default:
			if err := binary.Write(buf, binary.LittleEndian, value); err != nil {
				intlog.Errorf(context.TODO(), `%+v`, err)
				buf.Write(LeEncodeString(fmt.Sprintf("%v", value)))
			}
		}
	}
	return buf.Bytes()
}

// LeEncodeByLength 使用小端序将值编码为指定长度的字节切片。
// 如果编码结果长度小于指定长度，则在末尾补零；如果大于指定长度，则截断。
func LeEncodeByLength(length int, values ...interface{}) []byte {
	b := LeEncode(values...)
	if len(b) < length {
		b = append(b, make([]byte, length-len(b))...)
	} else if len(b) > length {
		b = b[0:length]
	}
	return b
}

// LeDecode 使用小端序将字节切片解码到指定的值中。
// 如果解码过程中发生错误，将返回包装后的错误信息。
func LeDecode(b []byte, values ...interface{}) error {
	var (
		err error
		buf = bytes.NewBuffer(b)
	)
	for i := 0; i < len(values); i++ {
		if err = binary.Read(buf, binary.LittleEndian, values[i]); err != nil {
			err = gerror.Wrap(err, `binary.Read failed`)
			return err
		}
	}
	return nil
}

// LeEncodeString 将字符串编码为字节切片。
func LeEncodeString(s string) []byte {
	return []byte(s)
}

// LeDecodeToString 将字节切片解码为字符串。
func LeDecodeToString(b []byte) string {
	return string(b)
}

// LeEncodeBool 将布尔值编码为字节切片。
// true编码为[1]，false编码为[0]。
func LeEncodeBool(b bool) []byte {
	if b {
		return []byte{1}
	} else {
		return []byte{0}
	}
}

// LeEncodeInt 将int类型的整数编码为字节切片。
// 根据整数大小自动选择合适的字节长度进行编码。
func LeEncodeInt(i int) []byte {
	if i <= math.MaxInt8 {
		return EncodeInt8(int8(i))
	} else if i <= math.MaxInt16 {
		return EncodeInt16(int16(i))
	} else if i <= math.MaxInt32 {
		return EncodeInt32(int32(i))
	} else {
		return EncodeInt64(int64(i))
	}
}

// LeEncodeUint 将uint类型的无符号整数编码为字节切片。
// 根据整数大小自动选择合适的字节长度进行编码。
func LeEncodeUint(i uint) []byte {
	if i <= math.MaxUint8 {
		return EncodeUint8(uint8(i))
	} else if i <= math.MaxUint16 {
		return EncodeUint16(uint16(i))
	} else if i <= math.MaxUint32 {
		return EncodeUint32(uint32(i))
	} else {
		return EncodeUint64(uint64(i))
	}
}

// LeEncodeInt8 将int8类型的整数编码为字节切片。
func LeEncodeInt8(i int8) []byte {
	return []byte{byte(i)}
}

// LeEncodeUint8 将uint8类型的无符号整数编码为字节切片。
func LeEncodeUint8(i uint8) []byte {
	return []byte{i}
}

// LeEncodeInt16 将int16类型的整数编码为小端序字节切片。
func LeEncodeInt16(i int16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(i))
	return b
}

// LeEncodeUint16 将uint16类型的无符号整数编码为小端序字节切片。
func LeEncodeUint16(i uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, i)
	return b
}

// LeEncodeInt32 将int32类型的整数编码为小端序字节切片。
func LeEncodeInt32(i int32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(i))
	return b
}

// LeEncodeUint32 将uint32类型的无符号整数编码为小端序字节切片。
func LeEncodeUint32(i uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
	return b
}

// LeEncodeInt64 将int64类型的整数编码为小端序字节切片。
func LeEncodeInt64(i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
}

// LeEncodeUint64 将uint64类型的无符号整数编码为小端序字节切片。
func LeEncodeUint64(i uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return b
}

// LeEncodeFloat32 将float32类型的浮点数编码为小端序字节切片。
func LeEncodeFloat32(f float32) []byte {
	bits := math.Float32bits(f)
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, bits)
	return b
}

// LeEncodeFloat64 将float64类型的浮点数编码为小端序字节切片。
func LeEncodeFloat64(f float64) []byte {
	bits := math.Float64bits(f)
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, bits)
	return b
}

// LeDecodeToInt 将字节切片解码为int类型的整数。
// 根据字节切片的长度自动选择合适的解码方式。
func LeDecodeToInt(b []byte) int {
	if len(b) < 2 {
		return int(LeDecodeToUint8(b))
	} else if len(b) < 3 {
		return int(LeDecodeToUint16(b))
	} else if len(b) < 5 {
		return int(LeDecodeToUint32(b))
	} else {
		return int(LeDecodeToUint64(b))
	}
}

// LeDecodeToUint 将字节切片解码为uint类型的无符号整数。
// 根据字节切片的长度自动选择合适的解码方式。
func LeDecodeToUint(b []byte) uint {
	if len(b) < 2 {
		return uint(LeDecodeToUint8(b))
	} else if len(b) < 3 {
		return uint(LeDecodeToUint16(b))
	} else if len(b) < 5 {
		return uint(LeDecodeToUint32(b))
	} else {
		return uint(LeDecodeToUint64(b))
	}
}

// LeDecodeToBool 将字节切片解码为布尔值。
// 如果切片为空或全为零值，返回false，否则返回true。
func LeDecodeToBool(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	if bytes.Equal(b, make([]byte, len(b))) {
		return false
	}
	return true
}

// LeDecodeToInt8 将字节切片解码为int8类型的整数。
// 如果输入切片为空，将会panic。
func LeDecodeToInt8(b []byte) int8 {
	if len(b) == 0 {
		panic(`empty slice given`)
	}
	return int8(b[0])
}

// LeDecodeToUint8 将字节切片解码为uint8类型的无符号整数。
// 如果输入切片为空，将会panic。
func LeDecodeToUint8(b []byte) uint8 {
	if len(b) == 0 {
		panic(`empty slice given`)
	}
	return b[0]
}

// LeDecodeToInt16 将小端序字节切片解码为int16类型的整数。
func LeDecodeToInt16(b []byte) int16 {
	return int16(binary.LittleEndian.Uint16(LeFillUpSize(b, 2)))
}

// LeDecodeToUint16 将小端序字节切片解码为uint16类型的无符号整数。
func LeDecodeToUint16(b []byte) uint16 {
	return binary.LittleEndian.Uint16(LeFillUpSize(b, 2))
}

// LeDecodeToInt32 将小端序字节切片解码为int32类型的整数。
func LeDecodeToInt32(b []byte) int32 {
	return int32(binary.LittleEndian.Uint32(LeFillUpSize(b, 4)))
}

// LeDecodeToUint32 将小端序字节切片解码为uint32类型的无符号整数。
func LeDecodeToUint32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(LeFillUpSize(b, 4))
}

// LeDecodeToInt64 将小端序字节切片解码为int64类型的整数。
func LeDecodeToInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(LeFillUpSize(b, 8)))
}

// LeDecodeToUint64 将小端序字节切片解码为uint64类型的无符号整数。
func LeDecodeToUint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(LeFillUpSize(b, 8))
}

// LeDecodeToFloat32 将小端序字节切片解码为float32类型的浮点数。
func LeDecodeToFloat32(b []byte) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(LeFillUpSize(b, 4)))
}

// LeDecodeToFloat64 将小端序字节切片解码为float64类型的浮点数。
func LeDecodeToFloat64(b []byte) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(LeFillUpSize(b, 8)))
}

// LeFillUpSize 使用小端序将字节切片b填充到指定长度l。
// 注意：该函数会创建一个新的字节切片并复制原始数据，以避免修改原始参数。
func LeFillUpSize(b []byte, l int) []byte {
	if len(b) >= l {
		return b[:l]
	}
	c := make([]byte, l)
	copy(c, b)
	return c
}
