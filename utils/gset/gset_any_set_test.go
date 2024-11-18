package gset

import (
	"strings"
	"sync"
	"testing"
)

// 测试辅助函数：检查集合内容是否一致
func checkSetEqual(t *testing.T, got, want []interface{}, testName string) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("%s: length mismatch: got %d, want %d", testName, len(got), len(want))
		return
	}

	gotMap := make(map[interface{}]bool)
	for _, v := range got {
		gotMap[v] = true
	}

	for _, v := range want {
		if !gotMap[v] {
			t.Errorf("%s: missing value %v in result", testName, v)
		}
	}
}

// 基本功能测试
func TestNew(t *testing.T) {
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
			set := New(tt.safe...)
			if set == nil {
				t.Error("New() returned nil")
				return
			}
			if got := set.Size(); got != tt.wantSize {
				t.Errorf("Initial size = %d, want %d", got, tt.wantSize)
			}
		})
	}
}

// NewFrom测试
func TestNewFrom(t *testing.T) {
	tests := []struct {
		name  string
		items interface{}
		want  []interface{}
	}{
		{
			name:  "切片转集合",
			items: []interface{}{1, 2, 3},
			want:  []interface{}{1, 2, 3},
		},
		{
			name:  "去重测试",
			items: []interface{}{1, 1, 2, 2, 3},
			want:  []interface{}{1, 2, 3},
		},
		{
			name:  "混合类型",
			items: []interface{}{1, "2", true},
			want:  []interface{}{1, "2", true},
		},
		{
			name:  "空切片",
			items: []interface{}{},
			want:  []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := NewFrom(tt.items)
			got := set.Slice()
			checkSetEqual(t, got, tt.want, tt.name)
		})
	}
}

// Add 和 Contains 测试
func TestAddAndContains(t *testing.T) {
	tests := []struct {
		name     string
		items    []interface{}
		check    interface{}
		contains bool
	}{
		{
			name:     "添加数字",
			items:    []interface{}{1, 2, 3},
			check:    2,
			contains: true,
		},
		{
			name:     "添加字符串",
			items:    []interface{}{"a", "b", "c"},
			check:    "b",
			contains: true,
		},
		{
			name:     "检查不存在的元素",
			items:    []interface{}{1, 2, 3},
			check:    4,
			contains: false,
		},
		{
			name:     "添加nil",
			items:    []interface{}{nil},
			check:    nil,
			contains: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := New()
			set.Add(tt.items...)
			if got := set.Contains(tt.check); got != tt.contains {
				t.Errorf("Contains() = %v, want %v", got, tt.contains)
			}
		})
	}
}

// 并发安全性测试
func TestConcurrentAccess(t *testing.T) {
	set := New(true) // 创建并发安全的集合
	var wg sync.WaitGroup
	iterations := 1000
	goroutines := 10

	// 并发添加
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				set.Add(base + j)
			}
		}(i * iterations)
	}

	// 并发读取
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				set.Contains(j)
			}
		}()
	}

	wg.Wait()

	// 验证结果
	if set.Size() != goroutines*iterations {
		t.Errorf("Size = %d, want %d", set.Size(), goroutines*iterations)
	}
}

// String方法测试
func TestString(t *testing.T) {
	tests := []struct {
		name  string
		items []interface{}
		want  string
	}{
		{
			name:  "空集合",
			items: []interface{}{},
			want:  "[]",
		},
		{
			name:  "数字集合",
			items: []interface{}{1, 2, 3},
			want:  "[1,2,3]",
		},
		{
			name:  "字符串集合",
			items: []interface{}{"a", "b", "c"},
			want:  `["a","b","c"]`,
		},
		{
			name:  "混合类型",
			items: []interface{}{1, "2", true},
			want:  `[1,"2",true]`,
		},
		{
			name:  "包含nil",
			items: []interface{}{1, nil, "test"},
			want:  `[1,null,"test"]`,
		},
		{
			name:  "浮点数",
			items: []interface{}{1.23, 4.56},
			want:  `[1.23,4.56]`,
		},
		{
			name:  "布尔值",
			items: []interface{}{true, false},
			want:  `[false,true]`,
		},
		{
			name:  "特殊字符",
			items: []interface{}{`a"b`, `c\d`},
			want:  `["a\"b","c\\d"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := NewFrom(tt.items)
			got := set.String()
			// 移除所有空白字符后比较
			got = strings.ReplaceAll(got, " ", "")
			tt.want = strings.ReplaceAll(tt.want, " ", "")
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 集合操作测试
func TestSetOperations(t *testing.T) {
	tests := []struct {
		name      string
		set1Items []interface{}
		set2Items []interface{}
		wantUnion []interface{}
		wantInter []interface{}
	}{
		{
			name:      "基本集合操作",
			set1Items: []interface{}{1, 2, 3},
			set2Items: []interface{}{2, 3, 4},
			wantUnion: []interface{}{1, 2, 3, 4},
			wantInter: []interface{}{2, 3},
		},
		{
			name:      "空集合操作",
			set1Items: []interface{}{1, 2},
			set2Items: []interface{}{},
			wantUnion: []interface{}{1, 2},
			wantInter: []interface{}{},
		},
		{
			name:      "相同集合操作",
			set1Items: []interface{}{1, 2},
			set2Items: []interface{}{1, 2},
			wantUnion: []interface{}{1, 2},
			wantInter: []interface{}{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set1 := NewFrom(tt.set1Items)
			set2 := NewFrom(tt.set2Items)

			// 测试并集
			union := set1.Union(set2)
			gotUnion := union.Slice()
			checkSetEqual(t, gotUnion, tt.wantUnion, tt.name+" Union")

			// 测试交集
			inter := set1.Intersect(set2)
			gotInter := inter.Slice()
			checkSetEqual(t, gotInter, tt.wantInter, tt.name+" Intersect")
		})
	}
}

// 性能测试
func BenchmarkSet(b *testing.B) {
	// 添加元素性能测试
	b.Run("Add", func(b *testing.B) {
		set := New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			set.Add(i)
		}
	})

	// 包含检查性能测试
	b.Run("Contains", func(b *testing.B) {
		set := New()
		for i := 0; i < 1000; i++ {
			set.Add(i)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			set.Contains(i % 1000)
		}
	})

	// String()性能测试
	b.Run("String", func(b *testing.B) {
		set := New()
		for i := 0; i < 1000; i++ {
			set.Add(i)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = set.String()
		}
	})
}

// 内存压力测试
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory usage test in short mode")
	}

	set := New()
	items := 1000000

	// 添加大量元素
	for i := 0; i < items; i++ {
		set.Add(i)
	}

	// 验证大小
	if got := set.Size(); got != items {
		t.Errorf("Size = %d, want %d", got, items)
	}

	// 清理并验证内存释放
	set.Clear()
	if got := set.Size(); got != 0 {
		t.Errorf("Size after Clear = %d, want 0", got)
	}
}

// 边界条件测试
func TestEdgeCases(t *testing.T) {
	set := New()

	// nil测试
	t.Run("Nil operations", func(t *testing.T) {
		set.Add(nil)
		if !set.Contains(nil) {
			t.Error("Should contain nil after adding it")
		}
		set.Remove(nil)
		if set.Contains(nil) {
			t.Error("Should not contain nil after removing it")
		}
	})

	// 空字符串测试
	t.Run("Empty string", func(t *testing.T) {
		set.Add("")
		if !set.Contains("") {
			t.Error("Should contain empty string after adding it")
		}
	})

	// 特殊字符测试
	t.Run("Special characters", func(t *testing.T) {
		specialChars := []string{"\n", "\t", "\r", "\\", "\""}
		for _, char := range specialChars {
			set.Add(char)
			if !set.Contains(char) {
				t.Errorf("Should contain special character %q", char)
			}
		}
	})
}

// Iterator方法测试
func TestIterator(t *testing.T) {
	set := NewFrom([]interface{}{1, 2, 3, 4, 5})
	count := 0
	sum := 0

	set.Iterator(func(v interface{}) bool {
		count++
		sum += v.(int)
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
	set.Iterator(func(v interface{}) bool {
		count++
		return count < 3
	})

	if count != 3 {
		t.Errorf("Early termination count = %d, want 3", count)
	}
}

// TestRemove 测试Remove方法
func TestRemove(t *testing.T) {
	tests := []struct {
		name        string
		initial     []interface{}
		removeItems []interface{}
		want        []interface{}
	}{
		{
			name:        "移除单个元素",
			initial:     []interface{}{1, 2, 3},
			removeItems: []interface{}{2},
			want:        []interface{}{1, 3},
		},
		{
			name:        "移除多个元素",
			initial:     []interface{}{1, 2, 3, 4, 5},
			removeItems: []interface{}{2, 4},
			want:        []interface{}{1, 3, 5},
		},
		{
			name:        "移除不存在的元素",
			initial:     []interface{}{1, 2, 3},
			removeItems: []interface{}{4},
			want:        []interface{}{1, 2, 3},
		},
		{
			name:        "移除nil",
			initial:     []interface{}{1, nil, 3},
			removeItems: []interface{}{nil},
			want:        []interface{}{1, 3},
		},
		{
			name:        "移除空集合中的元素",
			initial:     []interface{}{},
			removeItems: []interface{}{1},
			want:        []interface{}{},
		},
		{
			name:        "移除重复元素",
			initial:     []interface{}{1, 2, 3},
			removeItems: []interface{}{2, 2, 2},
			want:        []interface{}{1, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := NewFrom(tt.initial)
			set.Remove(tt.removeItems...)
			got := set.Slice()
			checkSetEqual(t, got, tt.want, tt.name)
		})
	}
}

// TestRemoveAll 测试RemoveAll方法
func TestRemoveAll(t *testing.T) {
	tests := []struct {
		name string
		set1 []interface{}
		set2 []interface{}
		want []interface{}
	}{
		{
			name: "移除部分元素",
			set1: []interface{}{1, 2, 3, 4, 5},
			set2: []interface{}{2, 4},
			want: []interface{}{1, 3, 5},
		},
		{
			name: "移除所有元素",
			set1: []interface{}{1, 2, 3},
			set2: []interface{}{1, 2, 3},
			want: []interface{}{},
		},
		{
			name: "移除不存在的元素",
			set1: []interface{}{1, 2, 3},
			set2: []interface{}{4, 5, 6},
			want: []interface{}{1, 2, 3},
		},
		{
			name: "从空集合移除",
			set1: []interface{}{},
			set2: []interface{}{1, 2, 3},
			want: []interface{}{},
		},
		{
			name: "使用空集合移除",
			set1: []interface{}{1, 2, 3},
			set2: []interface{}{},
			want: []interface{}{1, 2, 3},
		},
		{
			name: "移除包含nil的集合",
			set1: []interface{}{1, nil, 3},
			set2: []interface{}{nil, 3},
			want: []interface{}{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set1 := NewFrom(tt.set1)
			set2 := NewFrom(tt.set2)
			set1.RemoveAll(set2)
			got := set1.Slice()
			checkSetEqual(t, got, tt.want, tt.name)
		})
	}
}

// TestRemoveConcurrent 测试并发移除操作
func TestRemoveConcurrent(t *testing.T) {
	set := New(true) // 创建并发安全的集合
	// 首先添加元素
	for i := 0; i < 1000; i++ {
		set.Add(i)
	}

	var wg sync.WaitGroup
	// 并发移除偶数
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := start; j < 1000; j += 10 {
				if j%2 == 0 {
					set.Remove(j)
				}
			}
		}(i)
	}

	// 并发移除奇数
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := start; j < 1000; j += 10 {
				if j%2 != 0 {
					set.Remove(j)
				}
			}
		}(i)
	}

	wg.Wait()

	// 验证结果
	if set.Size() != 0 {
		t.Errorf("After concurrent remove, size = %d, want 0", set.Size())
	}
}

// 性能测试
func BenchmarkRemove(b *testing.B) {
	// 准备测试数据
	items := make([]interface{}, 1000)
	for i := range items {
		items[i] = i
	}

	b.Run("单元素删除", func(b *testing.B) {
		set := NewFrom(items)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			set.Remove(i % 1000)
		}
	})

	b.Run("批量删除", func(b *testing.B) {
		set := NewFrom(items)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			removeItems := []interface{}{i % 1000, (i + 1) % 1000, (i + 2) % 1000}
			set.Remove(removeItems...)
		}
	})
}
