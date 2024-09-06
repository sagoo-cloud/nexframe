package gstr

import (
	"github.com/sagoo-cloud/nexframe/utils"
	"strings"
)

// Trim 从字符串的开头和结尾去除空白字符（或其他字符）
// 可选参数 characterMask 指定额外要去除的字符
func Trim(str string, characterMask ...string) string {
	return utils.Trim(str, characterMask...)
}

// TrimStr 从字符串的开头和结尾去除所有给定的 cut 字符串
// 注意：它不会去除开头或结尾的空白字符
func TrimStr(str string, cut string, count ...int) string {
	return TrimLeftStr(TrimRightStr(str, cut, count...), cut, count...)
}

// TrimLeft 从字符串的开头去除空白字符（或其他字符）
func TrimLeft(str string, characterMask ...string) string {
	trimChars := utils.DefaultTrimChars
	if len(characterMask) > 0 {
		trimChars += characterMask[0]
	}
	return strings.TrimLeft(str, trimChars)
}

// TrimLeftStr 从字符串的开头去除所有给定的 cut 字符串
// 注意：它不会去除开头的空白字符
func TrimLeftStr(str string, cut string, count ...int) string {
	maxCount := -1
	if len(count) > 0 && count[0] != -1 {
		maxCount = count[0]
	}
	for i := 0; maxCount == -1 || i < maxCount; i++ {
		if strings.HasPrefix(str, cut) {
			str = strings.TrimPrefix(str, cut)
		} else {
			break
		}
	}
	return str
}

// TrimRight 从字符串的结尾去除空白字符（或其他字符）
func TrimRight(str string, characterMask ...string) string {
	trimChars := utils.DefaultTrimChars
	if len(characterMask) > 0 {
		trimChars += characterMask[0]
	}
	return strings.TrimRight(str, trimChars)
}

// TrimRightStr 从字符串的结尾去除所有给定的 cut 字符串
// 注意：它不会去除结尾的空白字符
func TrimRightStr(str string, cut string, count ...int) string {
	maxCount := -1
	if len(count) > 0 && count[0] != -1 {
		maxCount = count[0]
	}
	for i := 0; maxCount == -1 || i < maxCount; i++ {
		if strings.HasSuffix(str, cut) {
			str = strings.TrimSuffix(str, cut)
		} else {
			break
		}
	}
	return str
}

// TrimAll 去除字符串中的所有指定字符
func TrimAll(str string, characterMask ...string) string {
	trimChars := utils.DefaultTrimChars
	if len(characterMask) > 0 {
		trimChars += characterMask[0]
	}
	var builder strings.Builder
	builder.Grow(len(str))
	for _, char := range str {
		if !strings.ContainsRune(trimChars, char) {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

// HasPrefix 测试字符串 s 是否以 prefix 开头
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix 测试字符串 s 是否以 suffix 结尾
func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}
