package gset

import (
	"bytes"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"github.com/sagoo-cloud/nexframe/utils/json"
	"github.com/sagoo-cloud/nexframe/utils/rwmutex"
	"sort"
	"sync"
)

// IntSet 存储不重复的整数集合
type IntSet struct {
	mu   rwmutex.RWMutex
	data map[int]struct{}
}

// 用于存储字符串构建器的对象池
var intBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 64))
	},
}

// NewIntSet 创建一个新的整数集合
// safe参数用于指定是否启用并发安全，默认为false
func NewIntSet(safe ...bool) *IntSet {
	return &IntSet{
		mu:   rwmutex.Create(safe...),
		data: make(map[int]struct{}, 8), // 预分配合适的初始容量
	}
}

// NewIntSetFrom 从整数切片创建新的集合
func NewIntSetFrom(items []int, safe ...bool) *IntSet {
	m := make(map[int]struct{}, len(items))
	for _, v := range items {
		m[v] = struct{}{}
	}
	return &IntSet{
		mu:   rwmutex.Create(safe...),
		data: m,
	}
}

// Iterator 使用回调函数只读遍历集合
func (set *IntSet) Iterator(f func(v int) bool) {
	set.mu.RLock()
	defer set.mu.RUnlock()

	for k := range set.data {
		if !f(k) {
			break
		}
	}
}

// Add 添加一个或多个整数到集合
func (set *IntSet) Add(items ...int) {
	if len(items) == 0 {
		return
	}

	set.mu.Lock()
	if set.data == nil {
		set.data = make(map[int]struct{}, len(items))
	}
	for _, v := range items {
		set.data[v] = struct{}{}
	}
	set.mu.Unlock()
}

// Contains 检查集合是否包含指定整数
func (set *IntSet) Contains(item int) bool {
	set.mu.RLock()
	_, exists := set.data[item]
	set.mu.RUnlock()
	return exists
}

// Remove 从集合中删除指定整数
func (set *IntSet) Remove(items ...int) {
	if len(items) == 0 {
		return
	}

	set.mu.Lock()
	for _, item := range items {
		delete(set.data, item)
	}
	set.mu.Unlock()
}

// Size 返回集合大小
func (set *IntSet) Size() int {
	set.mu.RLock()
	size := len(set.data)
	set.mu.RUnlock()
	return size
}

// Slice 返回包含所有元素的有序切片
func (set *IntSet) Slice() []int {
	set.mu.RLock()
	defer set.mu.RUnlock()

	slice := make([]int, 0, len(set.data))
	for k := range set.data {
		slice = append(slice, k)
	}
	sort.Ints(slice) // 保证输出顺序一致
	return slice
}

// String 返回集合的字符串表示
func (set *IntSet) String() string {
	if set == nil {
		return "[]"
	}

	set.mu.RLock()
	defer set.mu.RUnlock()

	// 获取缓存的buffer
	buf := intBufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		intBufferPool.Put(buf)
	}()

	buf.WriteByte('[')

	// 获取排序后的切片
	slice := make([]int, 0, len(set.data))
	for k := range set.data {
		slice = append(slice, k)
	}
	sort.Ints(slice)

	// 构建字符串
	for i, v := range slice {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(convert.String(v))
	}

	buf.WriteByte(']')
	return buf.String()
}

// Union 返回与其他集合的并集
func (set *IntSet) Union(others ...*IntSet) *IntSet {
	newSet := NewIntSet(set.mu.IsSafe())
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

// Intersect 返回与其他集合的交集
func (set *IntSet) Intersect(others ...*IntSet) *IntSet {
	if len(others) == 0 {
		return NewIntSet(set.mu.IsSafe())
	}

	newSet := NewIntSet(set.mu.IsSafe())
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

// Sum 计算集合中所有整数的和
func (set *IntSet) Sum() (sum int) {
	set.mu.RLock()
	for k := range set.data {
		sum += k
	}
	set.mu.RUnlock()
	return
}

// Pop 随机移除并返回一个元素
func (set *IntSet) Pop() int {
	set.mu.Lock()
	defer set.mu.Unlock()

	if len(set.data) == 0 {
		return 0
	}

	// 获取第一个元素
	for k := range set.data {
		delete(set.data, k)
		return k
	}
	return 0
}

// MarshalJSON 实现 json.Marshal 接口
func (set *IntSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(set.Slice())
}

// UnmarshalJSON 实现 json.Unmarshal 接口
func (set *IntSet) UnmarshalJSON(b []byte) error {
	if set == nil {
		return nil
	}

	set.mu.Lock()
	defer set.mu.Unlock()

	if set.data == nil {
		set.data = make(map[int]struct{})
	}

	var array []int
	if err := json.UnmarshalUseNumber(b, &array); err != nil {
		return err
	}

	for _, v := range array {
		set.data[v] = struct{}{}
	}
	return nil
}
