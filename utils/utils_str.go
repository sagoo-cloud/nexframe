// Package utils 提供通用的字符串处理工具函数
package utils

import (
	"bytes"
	"strings"
	"sync"
	"unicode"
)

var (
	// DefaultTrimChars 定义了默认的需要被去除的字符集
	DefaultTrimChars = string([]byte{
		'\t', // Tab
		'\v', // Vertical tab
		'\n', // New line (line feed)
		'\r', // Carriage return
		'\f', // New page
		' ',  // Ordinary space
		0x00, // NUL-byte
		0x85, // Delete
		0xA0, // Non-breaking space
	})

	// bytesBufferPool 用于复用 bytes.Buffer，减少内存分配
	bytesBufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

// IsLetterUpper 检查给定字节是否为大写字母
func IsLetterUpper(b byte) bool {
	return b >= 'A' && b <= 'Z'
}

// IsLetterLower 检查给定字节是否为小写字母
func IsLetterLower(b byte) bool {
	return b >= 'a' && b <= 'z'
}

// IsLetter 检查给定字节是否为字母
func IsLetter(b byte) bool {
	return IsLetterUpper(b) || IsLetterLower(b)
}

// IsNumeric 检查给定字符串是否为数字（包括浮点数）
func IsNumeric(s string) bool {
	if s == "" {
		return false
	}

	var dotCount int
	for i, ch := range s {
		switch {
		case ch == '-' && i == 0:
			continue
		case ch == '.':
			dotCount++
			if dotCount > 1 || i == 0 || i == len(s)-1 {
				return false
			}
		case ch < '0' || ch > '9':
			return false
		}
	}
	return true
}

// UcFirst 将字符串的第一个字母转换为大写
func UcFirst(s string) string {
	if s == "" {
		return s
	}
	if IsLetterLower(s[0]) {
		return string(s[0]-32) + s[1:]
	}
	return s
}

// ReplaceByMap 使用映射表替换字符串中的内容
func ReplaceByMap(origin string, replaces map[string]string) string {
	if origin == "" || len(replaces) == 0 {
		return origin
	}

	// 优化：按键长度降序排序，避免短字符串替换影响长字符串
	// 创建一个 Builder 来构建结果
	var builder strings.Builder
	builder.Grow(len(origin))

	result := origin
	for k, v := range replaces {
		result = strings.ReplaceAll(result, k, v)
	}
	return result
}

// RemoveSymbols 移除字符串中的所有符号，只保留数字和字母
func RemoveSymbols(s string) string {
	if s == "" {
		return s
	}

	var builder strings.Builder
	builder.Grow(len(s))

	for _, c := range s {
		switch {
		case c > 127:
			builder.WriteRune(c)
		case unicode.IsLetter(c) || unicode.IsDigit(c):
			builder.WriteRune(c)
		}
	}
	return builder.String()
}

// EqualFoldWithoutChars 比较两个字符串是否相等（忽略大小写和特殊字符）
func EqualFoldWithoutChars(s1, s2 string) bool {
	if s1 == "" && s2 == "" {
		return true
	}
	return strings.EqualFold(RemoveSymbols(s1), RemoveSymbols(s2))
}

// SplitAndTrim 分割字符串并对每个部分进行修剪
func SplitAndTrim(str, delimiter string, characterMask ...string) []string {
	if str == "" {
		return nil
	}

	// 预分配合适大小的切片
	parts := strings.Split(str, delimiter)
	result := make([]string, 0, len(parts))

	for _, v := range parts {
		if trimmed := Trim(v, characterMask...); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Trim 去除字符串首尾的空白字符或指定字符
func Trim(str string, characterMask ...string) string {
	if str == "" {
		return str
	}

	trimChars := DefaultTrimChars
	if len(characterMask) > 0 && characterMask[0] != "" {
		trimChars += characterMask[0]
	}
	return strings.Trim(str, trimChars)
}

// FormatCmdKey 格式化命令键名
func FormatCmdKey(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(strings.ReplaceAll(s, "_", "."))
}

// FormatEnvKey 格式化环境变量键名
func FormatEnvKey(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(strings.ReplaceAll(s, ".", "_"))
}

// StripSlashes 移除字符串中的转义字符
func StripSlashes(str string) string {
	if str == "" {
		return str
	}

	// 从对象池获取 buffer
	buf := bytesBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bytesBufferPool.Put(buf)

	var skip bool
	for i, char := range str {
		if skip {
			skip = false
			continue
		}
		if char == '\\' && i+1 < len(str) && str[i+1] == '\\' {
			skip = true
		}
		if !skip && char != '\\' {
			buf.WriteRune(char)
		}
	}
	return buf.String()
}
