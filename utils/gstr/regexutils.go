package gstr

import (
	"regexp"
	"strings"
)

// ExtractRegexMatch 提取正则表达式匹配的指定捕获组
// 参数：
//   - regex: 正则表达式字符串
//   - content: 要匹配的内容
//   - index: 捕获组索引
//
// 返回值：匹配的字符串和可能的错误
func ExtractRegexMatch(regex, content string, index int) (string, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}
	match := r.FindStringSubmatch(content)
	if len(match) > index {
		return match[index], nil
	}
	return "", nil
}

// ReplaceRegexMatch 使用新字符串替换所有正则表达式匹配项
// 参数：
//   - regex: 正则表达式字符串
//   - newStr: 用于替换的新字符串
//   - content: 要处理的内容
//
// 返回值：替换后的字符串和可能的错误
func ReplaceRegexMatch(regex, newStr, content string) (string, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}
	return r.ReplaceAllString(content, newStr), nil
}

// IsFloat 判断字符串是否为浮点数
// 参数：
//   - s: 要判断的字符串
//
// 返回值：如果是浮点数返回true，否则返回false
func IsFloat(s string) bool {
	pattern := `^[\+-]?(\d*\.\d+|\d+\.\d*)$`
	match, _ := regexp.MatchString(pattern, strings.TrimSpace(s))
	return match
}

// IsInt 判断字符串是否为整数
// 参数：
//   - s: 要判断的字符串
//
// 返回值：如果是整数返回true，否则返回false
func IsInt(s string) bool {
	pattern := `^[\+-]?\d+$`
	match, _ := regexp.MatchString(pattern, strings.TrimSpace(s))
	return match
}

// GetRootDomain 获取根域名
// 参数：
//   - url: 完整的URL或域名字符串
//
// 返回值：提取的根域名
func GetRootDomain(url string) string {
	pattern := `([a-z0-9--]{1,200})\.(com\.cn|net\.cn|org\.cn|edu\.cn|gov\.cn|ac\.cn|bj\.cn|tj\.cn|sh\.cn|cq\.cn|he\.cn|sx\.cn|nm\.cn|ln\.cn|jl\.cn|hl\.cn|js\.cn|zj\.cn|ah\.cn|fj\.cn|jx\.cn|sd\.cn|ha\.cn|hb\.cn|hn\.cn|gd\.cn|gx\.cn|hi\.cn|sc\.cn|gz\.cn|yn\.cn|xz\.cn|sn\.cn|gs\.cn|qh\.cn|nx\.cn|xj\.cn|tw\.cn|hk\.cn|mo\.cn|com|cn|net|org|xyz|top|icu|vip|club|shop|wang|info|online|tech|site|fun|cc|website|space|press|news|video|work|app|kim|link|today|live|mobi|beauty|group|co|design|pro|red|guru|pub|team|social|center|life|biz|game|world|city|cloud|company|cool|zone|earth|email|digital|plus|zone|today|store)(:?\d*)?$`
	domain, _ := ExtractRegexMatch(pattern, url, 0)
	if domain != "" {
		parts := strings.Split(domain, ":")
		return parts[0]
	}
	return ""
}
