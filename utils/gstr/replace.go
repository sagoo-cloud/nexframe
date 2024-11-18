package gstr

import (
	"github.com/sagoo-cloud/nexframe/utils"
	"strings"
	"sync"
)

// 用于存储临时字符串的对象池
var builderPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

// Replace 返回字符串origin的一个副本
// 其中search字符串被replace替换，区分大小写
// count参数控制替换次数，默认为-1表示替换所有
func Replace(origin, search, replace string, count ...int) string {
	// 当搜索串为空时，直接返回原字符串
	if search == "" {
		return origin
	}

	n := -1
	if len(count) > 0 {
		n = count[0]
	}
	return strings.Replace(origin, search, replace, n)
}

// ReplaceI 返回字符串origin的一个副本
// 其中search字符串被replace替换，不区分大小写
// count参数控制替换次数，默认为-1表示替换所有
func ReplaceI(origin, search, replace string, count ...int) string {
	// 参数验证
	if origin == "" || search == "" {
		return origin
	}

	n := -1
	if len(count) > 0 {
		n = count[0]
	}
	if n == 0 {
		return origin
	}

	// 获取Builder
	builder := builderPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset()
		builderPool.Put(builder)
	}()

	searchLower := strings.ToLower(search)
	originLower := strings.ToLower(origin)
	searchLen := len(search)

	lastPos := 0
	pos := 0

	// 使用Builder进行字符串拼接
	for {
		pos = strings.Index(originLower[lastPos:], searchLower)
		if pos == -1 || (n <= 0 && n != -1) {
			builder.WriteString(origin[lastPos:])
			break
		}
		pos += lastPos
		builder.WriteString(origin[lastPos:pos])
		builder.WriteString(replace)
		lastPos = pos + searchLen

		if n > 0 {
			n--
		}
	}

	return builder.String()
}

// ReplaceByArray 使用字符串数组进行替换，区分大小写
// array中的元素按照pairs处理：array[0]替换为array[1]，array[2]替换为array[3]，以此类推
func ReplaceByArray(origin string, array []string) string {
	if len(array) < 2 {
		return origin
	}

	builder := builderPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset()
		builderPool.Put(builder)
	}()

	builder.WriteString(origin)
	result := builder.String()

	for i := 0; i < len(array)-1; i += 2 {
		result = Replace(result, array[i], array[i+1])
	}
	return result
}

// ReplaceIByArray 使用字符串数组进行替换，不区分大小写
// array中的元素按照pairs处理：array[0]替换为array[1]，array[2]替换为array[3]，以此类推
func ReplaceIByArray(origin string, array []string) string {
	if len(array) < 2 {
		return origin
	}

	builder := builderPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset()
		builderPool.Put(builder)
	}()

	builder.WriteString(origin)
	result := builder.String()

	for i := 0; i < len(array)-1; i += 2 {
		result = ReplaceI(result, array[i], array[i+1])
	}
	return result
}

// ReplaceByMap 使用map进行替换，区分大小写
// replaces中的key会被替换为对应的value
func ReplaceByMap(origin string, replaces map[string]string) string {
	if len(replaces) == 0 {
		return origin
	}
	return utils.ReplaceByMap(origin, replaces)
}

// ReplaceIByMap 使用map进行替换，不区分大小写
// replaces中的key会被替换为对应的value
func ReplaceIByMap(origin string, replaces map[string]string) string {
	if len(replaces) == 0 {
		return origin
	}

	builder := builderPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset()
		builderPool.Put(builder)
	}()

	builder.WriteString(origin)
	result := builder.String()

	for k, v := range replaces {
		result = ReplaceI(result, k, v)
	}
	return result
}
