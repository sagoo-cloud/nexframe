package gstr

import (
	"strings"
)

// IsSubDomain 检查 subDomain 是否为 mainDomain 的子域名。
// 支持在 mainDomain 中使用 '*' 通配符。
func IsSubDomain(subDomain, mainDomain string) bool {
	// 移除端口号（如果存在）
	subDomain = removePort(subDomain)
	mainDomain = removePort(mainDomain)

	subParts := strings.Split(subDomain, ".")
	mainParts := strings.Split(mainDomain, ".")

	subLen := len(subParts)
	mainLen := len(mainParts)

	// 如果主域名比子域名长，检查前缀是否全为通配符
	if mainLen > subLen {
		for _, part := range mainParts[:mainLen-subLen] {
			if part != "*" {
				return false
			}
		}
		mainParts = mainParts[mainLen-subLen:]
		mainLen = len(mainParts)
	}

	// 如果主域名长度大于2且子域名比主域名长，则不是子域名
	if mainLen > 2 && subLen > mainLen {
		return false
	}

	// 从后向前比较域名部分
	for i := 1; i <= mainLen; i++ {
		if mainParts[mainLen-i] == "*" {
			continue
		}
		if mainParts[mainLen-i] != subParts[subLen-i] {
			return false
		}
	}

	return true
}

// removePort 从域名中移除端口号（如果存在）
func removePort(domain string) string {
	if i := strings.IndexByte(domain, ':'); i != -1 {
		return domain[:i]
	}
	return domain
}
