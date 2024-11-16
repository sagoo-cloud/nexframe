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

// BeEncode 使用大端序将一个或多个值编码为字节切片。
// 通过类型断言检查values中每个值的类型，并调用相应的转换函数进行字节转换。
// 支持常见变量类型的断言，对于不支持的类型，使用fmt.Sprintf将值转换为字符串后再转换为字节。
func BeEncode(values ...interface{}) []byte {
	buf := new(bytes.Buffer)
	for i := 0; i < len(values); i++ {
		if values[i] == nil {
			return buf.Bytes()
		}

		switch value := values[i].(type) {
		case int:
			buf.Write(BeEncodeInt(value))
		case int8:
			buf.Write(BeEncodeInt8(value))
		case int16:
			buf.Write(BeEncodeInt16(value))
		case int32:
			buf.Write(BeEncodeInt32(value))
		case int64:
			buf.Write(BeEncodeInt64(value))
		case uint:
			buf.Write(BeEncodeUint(value))
		case uint8:
			buf.Write(BeEncodeUint8(value))
		case uint16:
			buf.Write(BeEncodeUint16(value))
		case uint32:
			buf.Write(BeEncodeUint32(value))
		case uint64:
			buf.Write(BeEncodeUint64(value))
		case bool:
			buf.Write(BeEncodeBool(value))
		case string:
			buf.Write(BeEncodeString(value))
		case []byte:
			buf.Write(value)
		case float32:
			buf.Write(BeEncodeFloat32(value))
		case float64:
			buf.Write(BeEncodeFloat64(value))
		default:
			if err := binary.Write(buf, binary.BigEndian, value); err != nil {
				intlog.Errorf(context.TODO(), `%+v`, err)
				buf.Write(BeEncodeString(fmt.Sprintf("%v", value)))
			}
		}
	}
	return buf.Bytes()
}

// BeEncodeByLength 使用大端序将值编码为指定长度的字节切片。
// 如果编码结果长度小于指定长度，则在末尾补零；如果大于指定长度，则截断。
func BeEncodeByLength(length int, values ...interface{}) []byte {
	b := BeEncode(values...)
	if len(b) < length {
		b = append(b, make([]byte, length-len(b))...)
	} else if len(b) > length {
		b = b[0:length]
	}
	return b
}

// BeDecode 使用大端序将字节切片解码到指定的值中。
// 如果解码过程中发生错误，将返回包装后的错误信息。
func BeDecode(b []byte, values ...interface{}) error {
	var (
		err error
		buf = bytes.NewBuffer(b)
	)
	for i := 0; i < len(values); i++ {
		if err = binary.Read(buf, binary.BigEndian, values[i]); err != nil {
			err = gerror.Wrap(err, `binary.Read failed`)
			return err
		}
	}
	return nil
}

// BeEncodeString 将字符串编码为字节切片。
func BeEncodeString(s string) []byte {
	return []byte(s)
}

// BeDecodeToString 将字节切片解码为字符串。
func BeDecodeToString(b []byte) string {
	return string(b)
}

// BeEncodeBool 将布尔值编码为字节切片。
// true编码为[1]，false编码为[0]。
func BeEncodeBool(b bool) []byte {
	if b {
		return []byte{1}
	} else {
		return []byte{0}
	}
}

// BeEncodeInt 将int类型的整数编码为字节切片。
// 根据整数大小自动选择合适的字节长度进行编码。
func BeEncodeInt(i int) []byte {
	if i <= math.MaxInt8 {
		return BeEncodeInt8(int8(i))
	} else if i <= math.MaxInt16 {
		return BeEncodeInt16(int16(i))
	} else if i <= math.MaxInt32 {
		return BeEncodeInt32(int32(i))
	} else {
		return BeEncodeInt64(int64(i))
	}
}

// BeEncodeUint 将uint类型的无符号整数编码为字节切片。
// 根据整数大小自动选择合适的字节长度进行编码。
func BeEncodeUint(i uint) []byte {
	if i <= math.MaxUint8 {
		return BeEncodeUint8(uint8(i))
	} else if i <= math.MaxUint16 {
		return BeEncodeUint16(uint16(i))
	} else if i <= math.MaxUint32 {
		return BeEncodeUint32(uint32(i))
	} else {
		return BeEncodeUint64(uint64(i))
	}
}

// BeEncodeInt8 将int8类型的整数编码为字节切片。
func BeEncodeInt8(i int8) []byte {
	return []byte{byte(i)}
}

// BeEncodeUint8 将uint8类型的无符号整数编码为字节切片。
func BeEncodeUint8(i uint8) []byte {
	return []byte{i}
}

// BeEncodeInt16 将int16类型的整数编码为大端序字节切片。
func BeEncodeInt16(i int16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(i))
	return b
}

// BeEncodeUint16 将uint16类型的无符号整数编码为大端序字节切片。
func BeEncodeUint16(i uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	return b
}

// BeEncodeInt32 将int32类型的整数编码为大端序字节切片。
func BeEncodeInt32(i int32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i))
	return b
}

// BeEncodeUint32 将uint32类型的无符号整数编码为大端序字节切片。
func BeEncodeUint32(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return b
}

// BeEncodeInt64 将int64类型的整数编码为大端序字节切片。
func BeEncodeInt64(i int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return b
}

// BeEncodeUint64 将uint64类型的无符号整数编码为大端序字节切片。
func BeEncodeUint64(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

// BeEncodeFloat32 将float32类型的浮点数编码为大端序字节切片。
func BeEncodeFloat32(f float32) []byte {
	bits := math.Float32bits(f)
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, bits)
	return b
}

// BeEncodeFloat64 将float64类型的浮点数编码为大端序字节切片。
func BeEncodeFloat64(f float64) []byte {
	bits := math.Float64bits(f)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, bits)
	return b
}

// BeDecodeToInt 将字节切片解码为int类型的整数。
// 根据字节切片的长度自动选择合适的解码方式。
func BeDecodeToInt(b []byte) int {
	if len(b) < 2 {
		return int(BeDecodeToUint8(b))
	} else if len(b) < 3 {
		return int(BeDecodeToUint16(b))
	} else if len(b) < 5 {
		return int(BeDecodeToUint32(b))
	} else {
		return int(BeDecodeToUint64(b))
	}
}

// BeDecodeToUint 将字节切片解码为uint类型的无符号整数。
// 根据字节切片的长度自动选择合适的解码方式。
func BeDecodeToUint(b []byte) uint {
	if len(b) < 2 {
		return uint(BeDecodeToUint8(b))
	} else if len(b) < 3 {
		return uint(BeDecodeToUint16(b))
	} else if len(b) < 5 {
		return uint(BeDecodeToUint32(b))
	} else {
		return uint(BeDecodeToUint64(b))
	}
}

// BeDecodeToBool 将字节切片解码为布尔值。
// 如果切片为空或全为零值，返回false，否则返回true。
func BeDecodeToBool(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	if bytes.Equal(b, make([]byte, len(b))) {
		return false
	}
	return true
}

// BeDecodeToInt8 将字节切片解码为int8类型的整数。
// 如果输入切片为空，将会panic。
func BeDecodeToInt8(b []byte) int8 {
	if len(b) == 0 {
		panic(`empty slice given`)
	}
	return int8(b[0])
}

// BeDecodeToUint8 将字节切片解码为uint8类型的无符号整数。
// 如果输入切片为空，将会panic。
func BeDecodeToUint8(b []byte) uint8 {
	if len(b) == 0 {
		panic(`empty slice given`)
	}
	return b[0]
}

// BeDecodeToInt16 将大端序字节切片解码为int16类型的整数。
func BeDecodeToInt16(b []byte) int16 {
	return int16(binary.BigEndian.Uint16(BeFillUpSize(b, 2)))
}

// BeDecodeToUint16 将大端序字节切片解码为uint16类型的无符号整数。
func BeDecodeToUint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(BeFillUpSize(b, 2))
}

// BeDecodeToInt32 将大端序字节切片解码为int32类型的整数。
func BeDecodeToInt32(b []byte) int32 {
	return int32(binary.BigEndian.Uint32(BeFillUpSize(b, 4)))
}

// BeDecodeToUint32 将大端序字节切片解码为uint32类型的无符号整数。
func BeDecodeToUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(BeFillUpSize(b, 4))
}

// BeDecodeToInt64 将大端序字节切片解码为int64类型的整数。
func BeDecodeToInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(BeFillUpSize(b, 8)))
}

// BeDecodeToUint64 将大端序字节切片解码为uint64类型的无符号整数。
func BeDecodeToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(BeFillUpSize(b, 8))
}

// BeDecodeToFloat32 将大端序字节切片解码为float32类型的浮点数。
func BeDecodeToFloat32(b []byte) float32 {
	return math.Float32frombits(binary.BigEndian.Uint32(BeFillUpSize(b, 4)))
}

// BeDecodeToFloat64 将大端序字节切片解码为float64类型的浮点数。
func BeDecodeToFloat64(b []byte) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(BeFillUpSize(b, 8)))
}

// BeFillUpSize 使用大端序将字节切片b填充到指定长度l。
// 注意：该函数会创建一个新的字节切片并复制原始数据，以避免修改原始参数。
func BeFillUpSize(b []byte, l int) []byte {
	if len(b) >= l {
		return b[:l]
	}
	c := make([]byte, l)
	copy(c[l-len(b):], b)
	return c
}
