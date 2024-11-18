package gset

import (
	"encoding/json"
	"sync"
	"testing"
)

// 测试辅助函数：比较两个整数切片是否相等
func sliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestNewIntSet 测试集合创建
func TestNewIntSet(t *testing.T) {
	tests := []struct {
		name     string
		safe     []bool
		wantSize int
	}{
		{"默认非并发安全", nil, 0},
		{"显式非并发安全", []bool{false}, 0},
		{"并发安全", []bool{true}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := NewIntSet(tt.safe...)
			if set == nil {
				t.Error("NewIntSet() returned nil")
				return
			}
			if got := set.Size(); got != tt.wantSize {
				t.Errorf("Initial size = %d, want %d", got, tt.wantSize)
			}
		})
	}
}

// TestNewIntSetFrom 测试从切片创建集合
func TestNewIntSetFrom(t *testing.T) {
	tests := []struct {
		name     string
		items    []int
		want     []int
		wantSize int
	}{
		{
			name:     "空切片",
			items:    []int{},
			want:     []int{},
			wantSize: 0,
		},
		{
			name:     "有序数字",
			items:    []int{1, 2, 3},
			want:     []int{1, 2, 3},
			wantSize: 3,
		},
		{
			name:     "乱序数字",
			items:    []int{3, 1, 2},
			want:     []int{1, 2, 3},
			wantSize: 3,
		},
		{
			name:     "重复数字",
			items:    []int{1, 1, 2, 2, 3},
			want:     []int{1, 2, 3},
			wantSize: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := NewIntSetFrom(tt.items)
			if got := set.Size(); got != tt.wantSize {
				t.Errorf("Size() = %v, want %v", got, tt.wantSize)
			}

			slice := set.Slice()
			if !sliceEqual(slice, tt.want) {
				t.Errorf("Slice() = %v, want %v", slice, tt.want)
			}
		})
	}
}

// TestIntSetAdd 测试添加操作
func TestIntSetAdd(t *testing.T) {
	tests := []struct {
		name  string
		items []int
		add   []int
		want  []int
	}{
		{
			name:  "添加到空集合",
			items: []int{},
			add:   []int{1, 2, 3},
			want:  []int{1, 2, 3},
		},
		{
			name:  "添加重复元素",
			items: []int{1, 2},
			add:   []int{2, 3},
			want:  []int{1, 2, 3},
		},
		{
			name:  "添加到nil集合",
			items: nil,
			add:   []int{1},
			want:  []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := NewIntSetFrom(tt.items)
			set.Add(tt.add...)

			if got := set.Slice(); !sliceEqual(got, tt.want) {
				t.Errorf("After Add, got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIntSetRemove 测试删除操作
func TestIntSetRemove(t *testing.T) {
	tests := []struct {
		name   string
		items  []int
		remove []int
		want   []int
	}{
		{
			name:   "删除存在的元素",
			items:  []int{1, 2, 3},
			remove: []int{2},
			want:   []int{1, 3},
		},
		{
			name:   "删除不存在的元素",
			items:  []int{1, 2, 3},
			remove: []int{4},
			want:   []int{1, 2, 3},
		},
		{
			name:   "删除多个元素",
			items:  []int{1, 2, 3, 4},
			remove: []int{1, 3},
			want:   []int{2, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := NewIntSetFrom(tt.items)
			for _, item := range tt.remove {
				set.Remove(item)
			}

			if got := set.Slice(); !sliceEqual(got, tt.want) {
				t.Errorf("After Remove, got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIntSetContains 测试包含操作
func TestIntSetContains(t *testing.T) {
	set := NewIntSetFrom([]int{1, 2, 3})
	tests := []struct {
		name string
		item int
		want bool
	}{
		{"存在的元素", 2, true},
		{"不存在的元素", 4, false},
		{"边界值", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := set.Contains(tt.item); got != tt.want {
				t.Errorf("Contains(%v) = %v, want %v", tt.item, got, tt.want)
			}
		})
	}
}

// TestIntSetString 测试字符串表示
func TestIntSetString(t *testing.T) {
	tests := []struct {
		name  string
		items []int
		want  string
	}{
		{"空集合", []int{}, "[]"},
		{"单元素", []int{1}, "[1]"},
		{"多元素", []int{1, 2, 3}, "[1,2,3]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := NewIntSetFrom(tt.items)
			if got := set.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIntSetUnion 测试并集操作
func TestIntSetUnion(t *testing.T) {
	tests := []struct {
		name string
		set1 []int
		set2 []int
		want []int
	}{
		{
			name: "不相交集合",
			set1: []int{1, 2},
			set2: []int{3, 4},
			want: []int{1, 2, 3, 4},
		},
		{
			name: "有交集",
			set1: []int{1, 2, 3},
			set2: []int{2, 3, 4},
			want: []int{1, 2, 3, 4},
		},
		{
			name: "空集合",
			set1: []int{1, 2},
			set2: []int{},
			want: []int{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set1 := NewIntSetFrom(tt.set1)
			set2 := NewIntSetFrom(tt.set2)
			union := set1.Union(set2)

			if got := union.Slice(); !sliceEqual(got, tt.want) {
				t.Errorf("Union = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIntSetIntersect 测试交集操作
func TestIntSetIntersect(t *testing.T) {
	tests := []struct {
		name string
		set1 []int
		set2 []int
		want []int
	}{
		{
			name: "有交集",
			set1: []int{1, 2, 3},
			set2: []int{2, 3, 4},
			want: []int{2, 3},
		},
		{
			name: "无交集",
			set1: []int{1, 2},
			set2: []int{3, 4},
			want: []int{},
		},
		{
			name: "空集合",
			set1: []int{1, 2},
			set2: []int{},
			want: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set1 := NewIntSetFrom(tt.set1)
			set2 := NewIntSetFrom(tt.set2)
			intersect := set1.Intersect(set2)

			if got := intersect.Slice(); !sliceEqual(got, tt.want) {
				t.Errorf("Intersect = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIntSetJSON 测试JSON序列化和反序列化
func TestIntSetJSON(t *testing.T) {
	tests := []struct {
		name  string
		items []int
	}{
		{"空集合", []int{}},
		{"单元素", []int{1}},
		{"多元素", []int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := NewIntSetFrom(tt.items)

			// 测试序列化
			data, err := json.Marshal(set)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			// 测试反序列化
			newSet := NewIntSet()
			if err := json.Unmarshal(data, newSet); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			// 验证结果
			if !sliceEqual(set.Slice(), newSet.Slice()) {
				t.Errorf("JSON roundtrip: got %v, want %v", newSet.Slice(), set.Slice())
			}
		})
	}
}

// TestIntSetConcurrent 测试并发操作
func TestIntSetConcurrent(t *testing.T) {
	set := NewIntSet(true) // 启用并发安全
	var wg sync.WaitGroup
	n := 1000

	// 并发添加
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			set.Add(val)
		}(i)
	}

	// 并发查询
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			set.Contains(val)
		}(i)
	}

	wg.Wait()
	if set.Size() != n {
		t.Errorf("After concurrent operations, size = %d, want %d", set.Size(), n)
	}
}

// TestIntSetIterator 测试迭代器
func TestIntSetIterator(t *testing.T) {
	set := NewIntSetFrom([]int{1, 2, 3, 4, 5})
	sum := 0
	count := 0

	set.Iterator(func(v int) bool {
		sum += v
		count++
		return true
	})

	if count != 5 {
		t.Errorf("Iterator count = %d, want 5", count)
	}
	if sum != 15 {
		t.Errorf("Iterator sum = %d, want 15", sum)
	}

	// 测试提前终止
	count = 0
	set.Iterator(func(v int) bool {
		count++
		return count < 3
	})

	if count != 3 {
		t.Errorf("Early termination count = %d, want 3", count)
	}
}

// BenchmarkIntSetOperations 性能测试
func BenchmarkIntSetOperations(b *testing.B) {
	// 添加操作性能
	b.Run("Add", func(b *testing.B) {
		set := NewIntSet()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			set.Add(i)
		}
	})

	// 查找操作性能
	b.Run("Contains", func(b *testing.B) {
		set := NewIntSet()
		for i := 0; i < 1000; i++ {
			set.Add(i)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			set.Contains(i % 1000)
		}
	})

	// 并集操作性能
	b.Run("Union", func(b *testing.B) {
		set1 := NewIntSetFrom([]int{1, 2, 3})
		set2 := NewIntSetFrom([]int{3, 4, 5})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			set1.Union(set2)
		}
	})
}
