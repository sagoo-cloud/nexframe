package gset

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
	"sync"
)

// StrSet 存储不重复的字符串集合
type StrSet struct {
	mu   sync.RWMutex
	data map[string]struct{}
	safe bool
}

// bufferPool 用于复用字符串构建器
var strBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 64))
	},
}

// NewStrSet 创建一个新的字符串集合
// safe参数用于指定是否启用并发安全，默认为false
func NewStrSet(safe ...bool) *StrSet {
	return &StrSet{
		data: make(map[string]struct{}, 8), // 预分配初始容量
		safe: len(safe) > 0 && safe[0],
	}
}

// NewStrSetFrom 从字符串切片创建新的集合
func NewStrSetFrom(items []string, safe ...bool) *StrSet {
	m := make(map[string]struct{}, len(items))
	for _, v := range items {
		m[v] = struct{}{}
	}
	return &StrSet{
		data: m,
		safe: len(safe) > 0 && safe[0],
	}
}

// Add 添加一个或多个字符串到集合
func (set *StrSet) Add(items ...string) {
	if len(items) == 0 {
		return
	}

	if set.safe {
		set.mu.Lock()
		defer set.mu.Unlock()
	}

	if set.data == nil {
		set.data = make(map[string]struct{}, len(items))
	}
	for _, v := range items {
		set.data[v] = struct{}{}
	}
}

// Contains 检查集合是否包含指定字符串
func (set *StrSet) Contains(item string) bool {
	if set.safe {
		set.mu.RLock()
		defer set.mu.RUnlock()
	}
	_, exists := set.data[item]
	return exists
}

// ContainsI 检查集合是否包含指定字符串(不区分大小写)
func (set *StrSet) ContainsI(item string) bool {
	if set.safe {
		set.mu.RLock()
		defer set.mu.RUnlock()
	}

	for k := range set.data {
		if strings.EqualFold(k, item) {
			return true
		}
	}
	return false
}

// Remove 从集合中移除指定字符串
func (set *StrSet) Remove(items ...string) {
	if len(items) == 0 {
		return
	}

	if set.safe {
		set.mu.Lock()
		defer set.mu.Unlock()
	}

	for _, item := range items {
		delete(set.data, item)
	}
}

// Size 返回集合大小
func (set *StrSet) Size() int {
	if set.safe {
		set.mu.RLock()
		defer set.mu.RUnlock()
	}
	return len(set.data)
}

// Clear 清空集合
func (set *StrSet) Clear() {
	if set.safe {
		set.mu.Lock()
		defer set.mu.Unlock()
	}
	set.data = make(map[string]struct{})
}

// Slice 返回包含所有元素的有序切片
func (set *StrSet) Slice() []string {
	if set.safe {
		set.mu.RLock()
		defer set.mu.RUnlock()
	}

	slice := make([]string, 0, len(set.data))
	for k := range set.data {
		slice = append(slice, k)
	}
	sort.Strings(slice) // 保证输出顺序一致
	return slice
}

// String 返回集合的字符串表示
func (set *StrSet) String() string {
	if set == nil {
		return "[]"
	}

	if set.safe {
		set.mu.RLock()
		defer set.mu.RUnlock()
	}

	// 获取缓存的buffer
	buf := strBufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		strBufferPool.Put(buf)
	}()

	buf.WriteByte('[')

	// 获取排序后的切片以保证输出顺序一致
	items := make([]string, 0, len(set.data))
	for k := range set.data {
		items = append(items, k)
	}
	sort.Strings(items)

	for i, v := range items {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteByte('"')
		buf.WriteString(escapeString(v))
		buf.WriteByte('"')
	}

	buf.WriteByte(']')
	return buf.String()
}

// escapeString 转义字符串中的特殊字符
func escapeString(s string) string {
	var buf bytes.Buffer
	for _, c := range s {
		switch c {
		case '\\', '"':
			buf.WriteByte('\\')
			buf.WriteRune(c)
		case '\n':
			buf.WriteString("\\n")
		case '\r':
			buf.WriteString("\\r")
		case '\t':
			buf.WriteString("\\t")
		default:
			buf.WriteRune(c)
		}
	}
	return buf.String()
}

// Union 返回与其他集合的并集
func (set *StrSet) Union(others ...*StrSet) *StrSet {
	newSet := NewStrSet(set.safe)

	if set.safe {
		set.mu.RLock()
		defer set.mu.RUnlock()
	}

	for k := range set.data {
		newSet.data[k] = struct{}{}
	}

	for _, other := range others {
		if other == nil {
			continue
		}
		if other.safe {
			other.mu.RLock()
		}
		for k := range other.data {
			newSet.data[k] = struct{}{}
		}
		if other.safe {
			other.mu.RUnlock()
		}
	}

	return newSet
}

// Intersect 返回与其他集合的交集
func (set *StrSet) Intersect(others ...*StrSet) *StrSet {
	newSet := NewStrSet(set.safe)
	if len(others) == 0 {
		return newSet
	}

	if set.safe {
		set.mu.RLock()
		defer set.mu.RUnlock()
	}

	for k := range set.data {
		isInAll := true
		for _, other := range others {
			if other == nil {
				isInAll = false
				break
			}
			if other.safe {
				other.mu.RLock()
			}
			if _, ok := other.data[k]; !ok {
				if other.safe {
					other.mu.RUnlock()
				}
				isInAll = false
				break
			}
			if other.safe {
				other.mu.RUnlock()
			}
		}
		if isInAll {
			newSet.data[k] = struct{}{}
		}
	}

	return newSet
}

// MarshalJSON 实现 json.Marshaler 接口
func (set *StrSet) MarshalJSON() ([]byte, error) {
	if set.safe {
		set.mu.RLock()
		defer set.mu.RUnlock()
	}
	return json.Marshal(set.Slice())
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (set *StrSet) UnmarshalJSON(data []byte) error {
	if set.safe {
		set.mu.Lock()
		defer set.mu.Unlock()
	}

	var items []string
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	if set.data == nil {
		set.data = make(map[string]struct{}, len(items))
	}

	for _, item := range items {
		set.data[item] = struct{}{}
	}
	return nil
}

// Iterator 使用回调函数只读遍历集合
// 如果回调函数返回false则停止遍历
func (set *StrSet) Iterator(f func(v string) bool) {
	if set == nil {
		return
	}

	if set.safe {
		set.mu.RLock()
		defer set.mu.RUnlock()
	}

	// 为了保证遍历顺序稳定，先获取已排序的切片
	items := make([]string, 0, len(set.data))
	for item := range set.data {
		items = append(items, item)
	}
	sort.Strings(items)

	// 遍历排序后的切片
	for _, item := range items {
		if !f(item) {
			break
		}
	}
}
