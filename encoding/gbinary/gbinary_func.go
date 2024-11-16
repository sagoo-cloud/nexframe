// Package gbinary 提供了基本类型和字节切片之间的编码和解码功能。
// 支持大端序和小端序两种字节序，提供了对整数、浮点数、布尔值、字符串等基本类型的编解码操作。
// 此外还包含了对二进制位操作的实验性功能。
//
// 基本用法:
//
//	// 编码示例
//	bytes := gbinary.Encode(123, "hello", true)
//
//	// 解码示例
//	var (
//	    num int
//	    str string
//	    boo bool
//	)
//	err := gbinary.Decode(bytes, &num, &str, &boo)
package gbinary
