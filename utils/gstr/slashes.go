package gstr

import (
	"strings"
)

// 默认需要转义的字符集
var defaultMetaChars = map[rune]struct{}{
	'.': {}, '+': {}, '\\': {}, '(': {}, '$': {},
	')': {}, '[': {}, '^': {}, ']': {}, '*': {},
	'?': {}, '{': {}, '}': {},
}

// QuoteMeta 返回一个转义后的字符串，对特殊字符添加反斜杠（\）
// 如果提供了自定义字符集chars，则只转义这些字符
// 否则使用默认字符集: .\+*?[^]($){}
// str: 需要处理的字符串
// chars: 可选的自定义字符集，只转义这些字符中包含的字符
func QuoteMeta(str string, chars ...string) string {
	if str == "" {
		return str
	}

	// 获取字符串构建器
	builder := builderPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset()
		builderPool.Put(builder)
	}()

	// 预分配合适的缓冲区大小
	builder.Grow(len(str) * 2)

	// 如果提供了自定义字符集且不为空
	if len(chars) > 0 {
		// 创建自定义字符集的查找表
		customChars := make(map[rune]struct{}, len(chars[0]))
		for _, c := range chars[0] {
			customChars[c] = struct{}{}
		}

		// 处理字符串
		for _, char := range str {
			if _, needQuote := customChars[char]; needQuote {
				builder.WriteRune('\\')
			}
			builder.WriteRune(char)
		}
		return builder.String()
	}

	// 使用默认字符集
	for _, char := range str {
		if _, needQuote := defaultMetaChars[char]; needQuote {
			builder.WriteRune('\\')
		}
		builder.WriteRune(char)
	}

	return builder.String()
}
