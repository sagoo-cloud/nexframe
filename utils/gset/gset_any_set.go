package gset

import (
	"bytes"
	"fmt"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"github.com/sagoo-cloud/nexframe/utils/gstr"
	"github.com/sagoo-cloud/nexframe/utils/rwmutex"
	"sort"
	"sync"
)

// Set 是一个通用的集合类型，存储不重复的items
type Set struct {
	mu   rwmutex.RWMutex
	data map[interface{}]struct{}
}

// bufferPool 用于复用字符串构建器
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// New 创建并返回一个新的集合
// safe参数指定是否启用并发安全，默认为false
func New(safe ...bool) *Set {
	return NewSet(safe...)
}

// NewSet 创建并返回一个新的集合
func NewSet(safe ...bool) *Set {
	return &Set{
		data: make(map[interface{}]struct{}),
		mu:   rwmutex.Create(safe...),
	}
}

// NewFrom 从给定的items创建一个新的集合
// items可以是任意类型的变量或切片
func NewFrom(items interface{}, safe ...bool) *Set {
	m := make(map[interface{}]struct{}, len(convert.Interfaces(items)))
	for _, v := range convert.Interfaces(items) {
		m[v] = struct{}{}
	}
	return &Set{
		data: m,
		mu:   rwmutex.Create(safe...),
	}
}

// Iterator 使用回调函数只读遍历集合
// 如果f返回true则继续遍历，返回false则停止
func (set *Set) Iterator(f func(v interface{}) bool) {
	set.mu.RLock()
	defer set.mu.RUnlock()

	for k := range set.data {
		if !f(k) {
			break
		}
	}
}

// Add 添加一个或多个元素到集合中
func (set *Set) Add(items ...interface{}) {
	if len(items) == 0 {
		return
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	if set.data == nil {
		set.data = make(map[interface{}]struct{}, len(items))
	}

	for _, v := range items {
		set.data[v] = struct{}{}
	}
}

// AddIfNotExist 如果元素不存在则添加并返回true，否则返回false
func (set *Set) AddIfNotExist(item interface{}) bool {
	if item == nil {
		return false
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	if set.data == nil {
		set.data = make(map[interface{}]struct{})
	}

	if _, exists := set.data[item]; !exists {
		set.data[item] = struct{}{}
		return true
	}
	return false
}

// Contains 检查集合是否包含指定元素
func (set *Set) Contains(item interface{}) bool {
	if set == nil {
		return false
	}

	set.mu.RLock()
	_, exists := set.data[item]
	set.mu.RUnlock()
	return exists
}

// Size 返回集合的大小
func (set *Set) Size() int {
	if set == nil {
		return 0
	}

	set.mu.RLock()
	length := len(set.data)
	set.mu.RUnlock()
	return length
}

// Clear 清空集合
func (set *Set) Clear() {
	set.mu.Lock()
	set.data = make(map[interface{}]struct{})
	set.mu.Unlock()
}

// Slice 返回包含集合所有元素的切片
func (set *Set) Slice() []interface{} {
	if set == nil {
		return nil
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	slice := make([]interface{}, 0, len(set.data))
	for item := range set.data {
		slice = append(slice, item)
	}
	return slice
}

// String 返回集合的字符串表示
// String 返回集合的字符串表示，实现类似 JSON.Marshal 的效果
func (set *Set) String() string {
	if set == nil {
		return "[]"
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	// 获取缓存的buffer
	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
	}()

	buf.WriteByte('[')
	first := true

	// 为了保证输出顺序一致，先收集所有元素
	items := make([]interface{}, 0, len(set.data))
	for k := range set.data {
		items = append(items, k)
	}

	// 排序，确保输出顺序一致
	sort.Slice(items, func(i, j int) bool {
		// 转换为字符串进行比较
		return fmt.Sprint(items[i]) < fmt.Sprint(items[j])
	})

	for _, item := range items {
		if first {
			first = false
		} else {
			buf.WriteByte(',')
		}

		switch v := item.(type) {
		case nil:
			buf.WriteString("null")
		case bool:
			if v {
				buf.WriteString("true")
			} else {
				buf.WriteString("false")
			}
		case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			buf.WriteString(fmt.Sprint(v))
		default:
			// 字符串和其他类型都用引号包围
			buf.WriteString(`"` + gstr.QuoteMeta(convert.String(v), `"\`) + `"`)
		}
	}

	buf.WriteByte(']')
	return buf.String()
}

// Union 返回当前集合与其他集合的并集
func (set *Set) Union(others ...*Set) *Set {
	newSet := NewSet(set.mu.IsSafe())
	newSet.Add(set.Slice()...)

	for _, other := range others {
		if other != nil {
			other.mu.RLock()
			for item := range other.data {
				newSet.data[item] = struct{}{}
			}
			other.mu.RUnlock()
		}
	}
	return newSet
}

// Intersect 返回当前集合与其他集合的交集
func (set *Set) Intersect(others ...*Set) *Set {
	if len(others) == 0 {
		return NewSet(set.mu.IsSafe())
	}

	newSet := NewSet(set.mu.IsSafe())
	set.mu.RLock()
	defer set.mu.RUnlock()

	// 使用第一个集合作为基准
	first := others[0]
	if first == nil {
		return newSet
	}

	first.mu.RLock()
	// 遍历当前集合中的元素
	for item := range set.data {
		if _, exists := first.data[item]; exists {
			exists = true
			// 检查是否在所有其他集合中都存在
			for _, other := range others[1:] {
				if other == nil {
					exists = false
					break
				}
				other.mu.RLock()
				if _, ok := other.data[item]; !ok {
					exists = false
					other.mu.RUnlock()
					break
				}
				other.mu.RUnlock()
			}
			if exists {
				newSet.data[item] = struct{}{}
			}
		}
	}
	first.mu.RUnlock()

	return newSet
}

// Remove 从集合中移除指定元素
func (set *Set) Remove(items ...interface{}) {
	if len(items) == 0 || set == nil {
		return
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	if set.data != nil {
		for _, item := range items {
			delete(set.data, item)
		}
	}
}

// RemoveAll 移除另一个集合中的所有元素
func (set *Set) RemoveAll(other *Set) {
	if set == nil || other == nil {
		return
	}

	other.mu.RLock()
	defer other.mu.RUnlock()

	set.mu.Lock()
	defer set.mu.Unlock()

	if set.data != nil {
		for item := range other.data {
			delete(set.data, item)
		}
	}
}
