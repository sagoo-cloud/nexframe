package gstructs

import (
	"reflect"
	"sync"

	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/tag"
)

// tagCacheKey 用于标识字段的缓存键
type tagCacheKey struct {
	pkg    string  // 包路径
	typ    string  // 类型名
	name   string  // 字段名
	offset uintptr // 字段偏移量
}

// 创建缓存键
func newTagCacheKey(field reflect.StructField) tagCacheKey {
	return tagCacheKey{
		pkg:    field.PkgPath,
		typ:    field.Type.String(),
		name:   field.Name,
		offset: field.Offset,
	}
}

// 字段缓存，用于存储已解析的标签信息
var tagCache = struct {
	sync.RWMutex
	m map[tagCacheKey]map[string]string
}{
	m: make(map[tagCacheKey]map[string]string),
}

// Tag 返回字段标签中与key关联的值
func (f *Field) Tag(key string) string {
	if f == nil || !f.Value.IsValid() {
		return ""
	}

	cacheKey := newTagCacheKey(f.Field)

	// 先查找缓存
	tagCache.RLock()
	if cachedTags, ok := tagCache.m[cacheKey]; ok {
		if value, exists := cachedTags[key]; exists {
			tagCache.RUnlock()
			return value
		}
	}
	tagCache.RUnlock()

	s := f.Field.Tag.Get(key)
	if s != "" {
		s = tag.Parse(s)

		// 更新缓存
		tagCache.Lock()
		if _, ok := tagCache.m[cacheKey]; !ok {
			tagCache.m[cacheKey] = make(map[string]string)
		}
		tagCache.m[cacheKey][key] = s
		tagCache.Unlock()
	}
	return s
}

// TagLookup 返回字段标签中与key关联的值和是否存在的标志
func (f *Field) TagLookup(key string) (value string, ok bool) {
	if f == nil || !f.Value.IsValid() {
		return "", false
	}

	value, ok = f.Field.Tag.Lookup(key)
	if ok && value != "" {
		value = tag.Parse(value)
	}
	return
}

// IsEmbedded 返回字段是否为嵌入字段
func (f *Field) IsEmbedded() bool {
	return f != nil && f.Field.Anonymous
}

// TagStr 返回字段的标签字符串
func (f *Field) TagStr() string {
	if f == nil {
		return ""
	}
	return string(f.Field.Tag)
}

// TagMap 返回字段的所有标签及其值的映射
func (f *Field) TagMap() map[string]string {
	if f == nil {
		return nil
	}

	data := ParseTag(f.TagStr())
	if len(data) == 0 {
		return nil
	}

	for k, v := range data {
		data[k] = utils.StripSlashes(tag.Parse(v))
	}
	return data
}

// IsExported 返回字段是否为导出字段
func (f *Field) IsExported() bool {
	return f != nil && f.Field.PkgPath == ""
}

// Name 返回字段名
func (f *Field) Name() string {
	if f == nil {
		return ""
	}
	return f.Field.Name
}

// Type 返回字段类型
func (f *Field) Type() Type {
	if f == nil {
		return Type{}
	}
	return Type{Type: f.Field.Type}
}

// Kind 返回字段值的Kind
func (f *Field) Kind() reflect.Kind {
	if f == nil || !f.Value.IsValid() {
		return reflect.Invalid
	}
	return f.Value.Kind()
}

// OriginalKind 返回字段值的原始Kind（解引用后）
func (f *Field) OriginalKind() reflect.Kind {
	if f == nil || !f.Value.IsValid() {
		return reflect.Invalid
	}

	reflectType := f.Value.Type()
	reflectKind := reflectType.Kind()

	for reflectKind == reflect.Ptr {
		reflectType = reflectType.Elem()
		reflectKind = reflectType.Kind()
	}

	return reflectKind
}

// OriginalValue 返回字段的原始值（解引用后）
func (f *Field) OriginalValue() reflect.Value {
	if f == nil || !f.Value.IsValid() {
		return reflect.Value{}
	}

	reflectValue := f.Value
	reflectType := reflectValue.Type()
	reflectKind := reflectType.Kind()

	for reflectKind == reflect.Ptr && !f.IsNil() {
		reflectValue = reflectValue.Elem()
		reflectKind = reflectValue.Type().Kind()
	}

	return reflectValue
}

// IsEmpty 检查字段值是否为空
func (f *Field) IsEmpty() bool {
	if f == nil || !f.Value.IsValid() {
		return true
	}

	switch f.Value.Kind() {
	case reflect.Bool:
		return !f.Value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return f.Value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return f.Value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return f.Value.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return f.Value.Complex() == 0
	case reflect.String:
		return f.Value.String() == ""
	case reflect.Array, reflect.Slice:
		return f.Value.Len() == 0
	case reflect.Map, reflect.Chan:
		return f.Value.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return f.Value.IsNil()
	default:
		return false
	}
}

// IsNil 检查字段值是否为nil
func (f *Field) IsNil(traceSource ...bool) bool {
	if f == nil || !f.Value.IsValid() {
		return true
	}

	switch f.Value.Kind() {
	case reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice:
		return f.Value.IsNil()
	default:
		return false
	}
}
