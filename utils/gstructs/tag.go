package gstructs

import (
	"reflect"
	"strconv"
	"sync"

	"github.com/sagoo-cloud/nexframe/utils/errors/gcode"
	"github.com/sagoo-cloud/nexframe/utils/errors/gerror"
	"github.com/sagoo-cloud/nexframe/utils/tag"
)

// tagParseCache 用于缓存已解析的标签
var tagParseCache = struct {
	sync.RWMutex
	m map[string]map[string]string
}{
	m: make(map[string]map[string]string),
}

// ParseTag 解析标签字符串为map
func ParseTag(tagStr string) map[string]string {
	if tagStr == "" {
		return nil
	}

	// 检查缓存
	tagParseCache.RLock()
	if cached, ok := tagParseCache.m[tagStr]; ok {
		tagParseCache.RUnlock()
		return cached
	}
	tagParseCache.RUnlock()

	data := make(map[string]string)
	var key string

	// 解析标签
	for tagStr != "" {
		// 跳过前导空格
		i := 0
		for i < len(tagStr) && tagStr[i] == ' ' {
			i++
		}
		tagStr = tagStr[i:]
		if tagStr == "" {
			break
		}

		// 扫描到冒号
		i = 0
		for i < len(tagStr) && tagStr[i] > ' ' && tagStr[i] != ':' && tagStr[i] != '"' && tagStr[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tagStr) || tagStr[i] != ':' || tagStr[i+1] != '"' {
			break
		}
		key = tagStr[:i]
		tagStr = tagStr[i+1:]

		// 扫描引号内的值
		i = 1
		for i < len(tagStr) && tagStr[i] != '"' {
			if tagStr[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tagStr) {
			break
		}

		quotedValue := tagStr[:i+1]
		tagStr = tagStr[i+1:]

		value, err := strconv.Unquote(quotedValue)
		if err != nil {
			panic(gerror.WrapCodef(gcode.CodeInvalidParameter, err, `error parsing tag "%s"`, tagStr))
		}
		data[key] = tag.Parse(value)
	}

	// 更新缓存
	if len(data) > 0 {
		tagParseCache.Lock()
		tagParseCache.m[tagStr] = data
		tagParseCache.Unlock()
	}

	return data
}

// TagFields 获取结构体的标签字段
func TagFields(pointer interface{}, priority []string) ([]Field, error) {
	if pointer == nil {
		return nil, gerror.New("input pointer is nil")
	}
	return getFieldValuesByTagPriority(pointer, priority, make(map[string]struct{}))
}

// TagMapName 获取结构体的标签名称映射
func TagMapName(pointer interface{}, priority []string) (map[string]string, error) {
	fields, err := TagFields(pointer, priority)
	if err != nil {
		return nil, err
	}

	tagMap := make(map[string]string, len(fields))
	for _, field := range fields {
		tagMap[field.TagValue] = field.Name()
	}
	return tagMap, nil
}

// TagMapField 获取结构体的标签字段映射
func TagMapField(object interface{}, priority []string) (map[string]Field, error) {
	fields, err := TagFields(object, priority)
	if err != nil {
		return nil, err
	}

	tagMap := make(map[string]Field, len(fields))
	for _, field := range fields {
		tagField := field
		tagMap[field.TagValue] = tagField
	}
	return tagMap, nil
}

// getFieldValuesByTagPriority 根据标签优先级获取字段值
func getFieldValuesByTagPriority(
	pointer interface{},
	priority []string,
	repeatedTagFilteringMap map[string]struct{},
) ([]Field, error) {
	fields, err := getFieldValues(pointer)
	if err != nil {
		return nil, err
	}

	tagFields := make([]Field, 0, len(fields))
	for _, field := range fields {
		// 只处理导出的字段
		if !field.IsExported() {
			continue
		}

		// 获取标签值
		tagValue := ""
		tagName := ""
		for _, p := range priority {
			tagName = p
			tagValue = field.Tag(p)
			if tagValue != "" && tagValue != "-" {
				break
			}
		}

		if tagValue != "" {
			// 过滤重复标签
			if _, ok := repeatedTagFilteringMap[tagValue]; ok {
				continue
			}
			tagField := field
			tagField.TagName = tagName
			tagField.TagValue = tagValue
			tagFields = append(tagFields, tagField)
		}

		// 处理嵌入字段
		if field.IsEmbedded() && field.OriginalKind() == reflect.Struct {
			subTagFields, err := getFieldValuesByTagPriority(
				field.Value,
				priority,
				repeatedTagFilteringMap,
			)
			if err != nil {
				return nil, err
			}
			tagFields = append(tagFields, subTagFields...)
		}
	}

	return tagFields, nil
}
