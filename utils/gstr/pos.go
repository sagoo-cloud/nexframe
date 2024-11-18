package gstr

import (
	"strings"
	"sync"
)

// 用于存储临时字符串的对象池，减少内存分配
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

// Pos 返回needle在haystack中从startOffset开始的第一次出现的位置
// 区分大小写，如果未找到返回-1
func Pos(haystack, needle string, startOffset ...int) int {
	// 参数校验
	length := len(haystack)
	offset := 0
	if len(startOffset) > 0 {
		offset = startOffset[0]
	}
	if length == 0 || offset > length || -offset > length {
		return -1
	}

	// 处理负偏移
	if offset < 0 {
		offset += length
	}

	// 使用strings.Index进行查找
	pos := strings.Index(haystack[offset:], needle)
	if pos == -1 {
		return -1
	}
	return pos + offset
}

// PosRune 功能与Pos相同，但将输入视为Unicode字符串
func PosRune(haystack, needle string, startOffset ...int) int {
	pos := Pos(haystack, needle, startOffset...)
	if pos < 3 {
		return pos
	}
	return len([]rune(haystack[:pos]))
}

// PosI 返回needle在haystack中从startOffset开始的第一次出现的位置
// 不区分大小写，如果未找到返回-1
func PosI(haystack, needle string, startOffset ...int) int {
	// 参数校验
	length := len(haystack)
	offset := 0
	if len(startOffset) > 0 {
		offset = startOffset[0]
	}
	if length == 0 || offset > length || -offset > length {
		return -1
	}

	// 处理负偏移
	if offset < 0 {
		offset += length
	}

	// 从对象池获取Builder
	builder := bufferPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset()
		bufferPool.Put(builder)
	}()

	// 转换为小写并查找
	builder.WriteString(strings.ToLower(haystack[offset:]))
	pos := strings.Index(builder.String(), strings.ToLower(needle))
	if pos == -1 {
		return -1
	}
	return pos + offset
}

// PosIRune 功能与PosI相同，但将输入视为Unicode字符串
func PosIRune(haystack, needle string, startOffset ...int) int {
	pos := PosI(haystack, needle, startOffset...)
	if pos < 3 {
		return pos
	}
	return len([]rune(haystack[:pos]))
}

// PosR 返回needle在haystack中从startOffset开始的最后一次出现的位置
// 区分大小写，如果未找到返回-1
func PosR(haystack, needle string, startOffset ...int) int {
	// 参数校验
	offset := 0
	if len(startOffset) > 0 {
		offset = startOffset[0]
	}
	pos, length := 0, len(haystack)
	if length == 0 || offset > length || -offset > length {
		return -1
	}

	// 处理偏移量
	if offset < 0 {
		haystack = haystack[:offset+length+1]
	} else {
		haystack = haystack[offset:]
	}

	// 查找最后出现位置
	pos = strings.LastIndex(haystack, needle)
	if offset > 0 && pos != -1 {
		pos += offset
	}
	return pos
}

// PosRRune 功能与PosR相同，但将输入视为Unicode字符串
func PosRRune(haystack, needle string, startOffset ...int) int {
	pos := PosR(haystack, needle, startOffset...)
	if pos < 3 {
		return pos
	}
	return len([]rune(haystack[:pos]))
}

// PosRI 返回needle在haystack中从startOffset开始的最后一次出现的位置
// 不区分大小写，如果未找到返回-1
func PosRI(haystack, needle string, startOffset ...int) int {
	// 参数校验
	offset := 0
	if len(startOffset) > 0 {
		offset = startOffset[0]
	}
	pos, length := 0, len(haystack)
	if length == 0 || offset > length || -offset > length {
		return -1
	}

	// 处理偏移量
	if offset < 0 {
		haystack = haystack[:offset+length+1]
	} else {
		haystack = haystack[offset:]
	}

	// 从对象池获取Builder
	builder := bufferPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset()
		bufferPool.Put(builder)
	}()

	// 转换为小写并查找
	builder.WriteString(strings.ToLower(haystack))
	pos = strings.LastIndex(builder.String(), strings.ToLower(needle))
	if offset > 0 && pos != -1 {
		pos += offset
	}
	return pos
}

// PosRIRune 功能与PosRI相同，但将输入视为Unicode字符串
func PosRIRune(haystack, needle string, startOffset ...int) int {
	pos := PosRI(haystack, needle, startOffset...)
	if pos < 3 {
		return pos
	}
	return len([]rune(haystack[:pos]))
}
