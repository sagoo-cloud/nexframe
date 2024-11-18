// Package gstr 提供字符串处理的工具函数
package gstr

import (
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"strings"
	"sync"
)

// 用于ChunkSplit的内存池，减少内存分配
var runePool = sync.Pool{
	New: func() interface{} {
		// 预分配一个合理大小的切片
		return make([]rune, 0, 1024)
	},
}

// Split 使用分隔符将字符串分割为切片
func Split(str, delimiter string) []string {
	if str == "" || delimiter == "" {
		return nil
	}
	return strings.Split(str, delimiter)
}

// SplitAndTrim 分割字符串并对结果进行修剪
func SplitAndTrim(str, delimiter string, characterMask ...string) []string {
	if str == "" || delimiter == "" {
		return nil
	}
	return utils.SplitAndTrim(str, delimiter, characterMask...)
}

// Join 使用分隔符连接字符串切片
func Join(array []string, sep string) string {
	if len(array) == 0 {
		return ""
	}
	return strings.Join(array, sep)
}

// JoinAny 连接任意类型的切片为字符串
func JoinAny(array interface{}, sep string) string {
	if array == nil {
		return ""
	}
	strs := convert.Strings(array)
	if len(strs) == 0 {
		return ""
	}
	return strings.Join(strs, sep)
}

// Explode PHP风格的字符串分割函数
func Explode(delimiter, str string) []string {
	if str == "" || delimiter == "" {
		return nil
	}
	return Split(str, delimiter)
}

// Implode PHP风格的字符串连接函数
func Implode(glue string, pieces []string) string {
	if len(pieces) == 0 {
		return ""
	}
	return strings.Join(pieces, glue)
}

// ChunkSplit 将字符串分割成小块
func ChunkSplit(body string, chunkLen int, end string) string {
	// 参数验证
	if body == "" {
		return ""
	}
	if chunkLen <= 0 {
		chunkLen = 76 // 使用默认值
	}
	if end == "" {
		end = "\r\n"
	}

	runes := []rune(body)
	endRunes := []rune(end)
	runesLen := len(runes)

	if runesLen <= 1 || runesLen < chunkLen {
		return body + end
	}

	// 从对象池获取切片
	ns := runePool.Get().([]rune)
	// 确保足够容量
	needed := runesLen + (runesLen/chunkLen+1)*len(endRunes)
	if cap(ns) < needed {
		ns = make([]rune, 0, needed)
	}
	ns = ns[:0]

	// 分块处理
	for i := 0; i < runesLen; i += chunkLen {
		end := i + chunkLen
		if end > runesLen {
			end = runesLen
		}
		ns = append(ns, runes[i:end]...)
		ns = append(ns, endRunes...)
	}

	result := string(ns)
	// 将切片放回对象池
	runePool.Put(ns)
	return result
}

// Fields 返回字符串中的单词切片
func Fields(str string) []string {
	if str == "" {
		return nil
	}
	return strings.Fields(str)
}
