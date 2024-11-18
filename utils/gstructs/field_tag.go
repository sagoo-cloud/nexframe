package gstructs

import (
	"strings"
	"sync"

	"github.com/sagoo-cloud/nexframe/utils/tag"
)

// tagValueCache 用于缓存已解析的标签值
var tagValueCache = struct {
	sync.RWMutex
	m map[string]string
}{
	m: make(map[string]string),
}

// TagJsonName 返回字段的json标签名
func (f *Field) TagJsonName() string {
	if f == nil {
		return ""
	}

	if jsonTag := f.Tag(tag.Json); jsonTag != "" {
		return strings.Split(jsonTag, ",")[0]
	}
	return ""
}

// TagDefault 返回字段的默认值标签
func (f *Field) TagDefault() string {
	if f == nil {
		return ""
	}

	// 先检查缓存
	cacheKey := f.Field.Name + "_default"
	tagValueCache.RLock()
	if v, ok := tagValueCache.m[cacheKey]; ok {
		tagValueCache.RUnlock()
		return v
	}
	tagValueCache.RUnlock()

	// 获取标签值
	v := f.Tag(tag.Default)
	if v == "" {
		v = f.Tag(tag.DefaultShort)
	}

	// 更新缓存
	if v != "" {
		tagValueCache.Lock()
		tagValueCache.m[cacheKey] = v
		tagValueCache.Unlock()
	}

	return v
}

// TagParam 返回字段的参数标签
func (f *Field) TagParam() string {
	if f == nil {
		return ""
	}

	v := f.Tag(tag.Param)
	if v == "" {
		v = f.Tag(tag.ParamShort)
	}
	return v
}

// TagValid 返回字段的验证标签
func (f *Field) TagValid() string {
	if f == nil {
		return ""
	}

	v := f.Tag(tag.Valid)
	if v == "" {
		v = f.Tag(tag.ValidShort)
	}
	return v
}

// TagDescription 返回字段的描述标签
func (f *Field) TagDescription() string {
	if f == nil {
		return ""
	}

	v := f.Tag(tag.Description)
	if v == "" {
		v = f.Tag(tag.DescriptionShort)
	}
	if v == "" {
		v = f.Tag(tag.DescriptionShort2)
	}
	return v
}

// TagSummary 返回字段的摘要标签
func (f *Field) TagSummary() string {
	if f == nil {
		return ""
	}

	v := f.Tag(tag.Summary)
	if v == "" {
		v = f.Tag(tag.SummaryShort)
	}
	if v == "" {
		v = f.Tag(tag.SummaryShort2)
	}
	return v
}

// TagAdditional 返回字段的附加标签
func (f *Field) TagAdditional() string {
	if f == nil {
		return ""
	}

	v := f.Tag(tag.Additional)
	if v == "" {
		v = f.Tag(tag.AdditionalShort)
	}
	return v
}

// TagExample 返回字段的示例标签
func (f *Field) TagExample() string {
	if f == nil {
		return ""
	}

	v := f.Tag(tag.Example)
	if v == "" {
		v = f.Tag(tag.ExampleShort)
	}
	return v
}

// TagIn 返回字段的in标签
func (f *Field) TagIn() string {
	if f == nil {
		return ""
	}
	return f.Tag(tag.In)
}

// TagPriorityName 返回字段的优先级标签名称
func (f *Field) TagPriorityName() string {
	if f == nil {
		return ""
	}

	name := f.Name()
	for _, tagName := range tag.StructTagPriority {
		if tagValue := f.Tag(tagName); tagValue != "" {
			name = tagValue
			break
		}
	}
	return name
}
