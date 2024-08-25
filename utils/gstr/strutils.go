package gstr

import (
	"fmt"
	"strings"
	"unicode"
)

// SubString 截取字符串
// 参数：
//   - str: 输入字符串
//   - start: 起始位置
//   - end: 结束位置（如果为-1，则截取到字符串末尾）
//
// 返回值：截取后的字符串
func SubString(str string, start, end int) string {
	runes := []rune(str)
	if end == -1 || end > len(runes) {
		return string(runes[start:])
	}
	return string(runes[start:end])
}

// FirstToUpper 将字符串的首字母转为大写
// 参数：
//   - str: 输入字符串
//
// 返回值：首字母大写的字符串
func FirstToUpper(str string) string {
	if str == "" {
		return ""
	}
	return strings.ToUpper(str[:1]) + str[1:]
}

// FirstToLower 将字符串的首字母转为小写
// 参数：
//   - str: 输入字符串
//
// 返回值：首字母小写的字符串
func FirstToLower(str string) string {
	if str == "" {
		return ""
	}
	return strings.ToLower(str[:1]) + str[1:]
}

// InStringArray 检查字符串是否在切片中
// 参数：
//   - needle: 要查找的字符串
//   - haystack: 字符串切片
//
// 返回值：如果找到返回true，否则返回false
func InStringArray(needle string, haystack []string) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}
	return false
}

// ResolveAddress 解析地址
// 参数：
//   - addr: 地址参数切片
//
// 返回值：解析后的地址字符串和可能的错误
func ResolveAddress(addr []string) (string, error) {
	switch len(addr) {
	case 0:
		return ":1122", nil
	case 1:
		return addr[0] + ":1122", nil
	case 2:
		return addr[0] + ":" + addr[1], nil
	default:
		return "", fmt.Errorf("参数过多：期望0-2个，实际得到%d个", len(addr))
	}
}

// ReplaceIndex 替换字符串中第n次出现的旧字符串
// 参数：
//   - s: 原字符串
//   - old: 要替换的旧字符串
//   - new: 新字符串
//   - n: 第几次出现（从0开始）
//
// 返回值：替换后的字符串
func ReplaceIndex(s, old, new string, n int) string {
	var builder strings.Builder
	parts := strings.SplitN(s, old, n+2)
	for i, part := range parts {
		if i > 0 {
			if i-1 == n {
				builder.WriteString(new)
			} else {
				builder.WriteString(old)
			}
		}
		builder.WriteString(part)
	}
	return builder.String()
}

// UnderLineName 将驼峰命名转换为下划线命名
// 参数：
//   - name: 驼峰命名的字符串
//
// 返回值：下划线命名的字符串
func UnderLineName(name string) string {
	var builder strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) && i > 0 {
			builder.WriteByte('_')
		}
		builder.WriteRune(unicode.ToLower(r))
	}
	return builder.String()
}

// HumpName 将下划线命名转换为驼峰命名
// 参数：
//   - name: 下划线命名的字符串
//
// 返回值：驼峰命名的字符串
func HumpName(name string) string {
	var builder strings.Builder
	toUpper := false
	for _, r := range name {
		if r == '_' {
			toUpper = true
		} else {
			if toUpper {
				builder.WriteRune(unicode.ToUpper(r))
				toUpper = false
			} else {
				builder.WriteRune(r)
			}
		}
	}
	return builder.String()
}
