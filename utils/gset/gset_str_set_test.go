package gset

import (
	"encoding/json"
	"sort"
	"strconv"
	"sync"
	"testing"
)

// 测试辅助函数：比较两个字符串切片是否相等
func strSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	// 复制切片并排序以确保比较顺序一致
	aCopy := make([]string, len(a))
	copy(aCopy, a)
	bCopy := make([]string, len(b))
	copy(bCopy, b)

	sort.Strings(aCopy)
	sort.Strings(bCopy)

	for i := range aCopy {
		if aCopy[i] != bCopy[i] {
			return false
		}
	}
	return true
}

// TestNewStrSet 测试集合创建
func TestNewStrSet(t *testing.T) {
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
			set := NewStrSet(tt.safe...)
			if set == nil {
				t.Error("NewStrSet() returned nil")
				return
			}
			if got := set.Size(); got != tt.wantSize {
				t.Errorf("Initial size = %d, want %d", got, tt.wantSize)
			}
		})
	}
}

// TestStrSetOperations 测试基本操作
func TestStrSetOperations(t *testing.T) {
	set := NewStrSet()

	// 测试添加
	t.Run("Add", func(t *testing.T) {
		set.Add("a", "b", "c")
		if got := set.Size(); got != 3 {
			t.Errorf("After Add, size = %d, want 3", got)
		}
		if !set.Contains("b") {
			t.Error("Add failed: set should contain 'b'")
		}
	})

	// 测试大小写不敏感的包含
	t.Run("ContainsI", func(t *testing.T) {
		set.Add("Hello")
		if !set.ContainsI("hELLo") {
			t.Error("ContainsI failed: should find 'hELLo'")
		}
	})

	// 测试移除
	t.Run("Remove", func(t *testing.T) {
		set.Remove("b")
		if set.Contains("b") {
			t.Error("Remove failed: set should not contain 'b'")
		}
	})

	// 测试清空
	t.Run("Clear", func(t *testing.T) {
		set.Clear()
		if set.Size() != 0 {
			t.Error("Clear failed: set should be empty")
		}
	})
}

// TestStrSetConcurrent 测试并发操作
func TestStrSetConcurrent(t *testing.T) {
	set := NewStrSet(true)
	var wg sync.WaitGroup
	n := 1000

	// 并发添加
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			set.Add(string(rune('a' + val%26)))
		}(i)
	}

	// 并发读取
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			set.Contains(string(rune('a' + val%26)))
		}(i)
	}

	wg.Wait()
	if set.Size() > 26 {
		t.Error("Concurrent operations failed: size larger than expected")
	}
}

// TestStrSetJSON 测试JSON序列化和反序列化
func TestStrSetJSON(t *testing.T) {
	set := NewStrSet()
	set.Add("hello", "world", "测试")

	// 测试序列化
	data, err := json.Marshal(set)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// 测试反序列化
	newSet := NewStrSet()
	if err := json.Unmarshal(data, newSet); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !strSliceEqual(set.Slice(), newSet.Slice()) {
		t.Error("JSON roundtrip failed: sets not equal")
	}
}

// TestStrSetUnionAndIntersect 测试并集和交集操作
func TestStrSetUnionAndIntersect(t *testing.T) {
	set1 := NewStrSetFrom([]string{"a", "b", "c"})
	set2 := NewStrSetFrom([]string{"b", "c", "d"})

	// 测试并集
	t.Run("Union", func(t *testing.T) {
		union := set1.Union(set2)
		expected := []string{"a", "b", "c", "d"}
		if !strSliceEqual(union.Slice(), expected) {
			t.Errorf("Union = %v, want %v", union.Slice(), expected)
		}
	})

	// 测试交集
	t.Run("Intersect", func(t *testing.T) {
		intersect := set1.Intersect(set2)
		expected := []string{"b", "c"}
		if !strSliceEqual(intersect.Slice(), expected) {
			t.Errorf("Intersect = %v, want %v", intersect.Slice(), expected)
		}
	})
}

// TestStrSetString 测试字符串表示
func TestStrSetString(t *testing.T) {
	tests := []struct {
		name  string
		items []string
		want  string
	}{
		{"空集合", []string{}, "[]"},
		{"单元素", []string{"test"}, `["test"]`},
		{"多元素", []string{"a", "b", "c"}, `["a","b","c"]`},
		{"特殊字符", []string{`a"b`, `c\d`}, `["a\"b","c\\d"]`},
		{"Unicode字符", []string{"你好", "世界"}, `["世界","你好"]`},
		{"转义字符", []string{"a\nb", "c\td"}, `["a\nb","c\td"]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := NewStrSetFrom(tt.items)
			got := set.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestStrSetEdgeCases 测试边界情况
func TestStrSetEdgeCases(t *testing.T) {
	// 测试nil集合
	t.Run("NilSet", func(t *testing.T) {
		var set *StrSet
		if str := set.String(); str != "[]" {
			t.Errorf("nil set String() = %v, want []", str)
		}
	})

	// 测试空字符串
	t.Run("EmptyString", func(t *testing.T) {
		set := NewStrSet()
		set.Add("")
		if !set.Contains("") {
			t.Error("set should contain empty string")
		}
	})

	// 测试大量数据
	t.Run("LargeDataset", func(t *testing.T) {
		set := NewStrSet()
		n := 10000
		for i := 0; i < n; i++ {
			set.Add("item" + strconv.Itoa(i))
		}
		if set.Size() != n {
			t.Errorf("size = %d, want %d", set.Size(), n)
		}
	})
}

// TestStrSetIterator 测试迭代器
func TestStrSetIterator(t *testing.T) {
	// 准备测试数据
	testData := []string{"a", "b", "c", "d", "e"}
	set := NewStrSetFrom(testData)

	// 测试完整遍历
	t.Run("Complete Iteration", func(t *testing.T) {
		visited := make([]string, 0)
		set.Iterator(func(v string) bool {
			visited = append(visited, v)
			return true
		})

		// 验证访问的元素数量
		if len(visited) != len(testData) {
			t.Errorf("Iterator visited %d items, want %d", len(visited), len(testData))
		}

		// 验证访问的元素是否正确
		sort.Strings(visited)
		expected := make([]string, len(testData))
		copy(expected, testData)
		sort.Strings(expected)
		if !strSliceEqual(visited, expected) {
			t.Errorf("Iterator visited items %v, want %v", visited, expected)
		}
	})

	// 测试提前终止遍历
	t.Run("Early Termination", func(t *testing.T) {
		count := 0
		set.Iterator(func(v string) bool {
			count++
			return count < 3 // 只遍历前3个元素
		})

		if count != 3 {
			t.Errorf("Iterator early termination failed: visited %d items, want 3", count)
		}
	})

	// 测试空集合
	t.Run("Empty Set", func(t *testing.T) {
		emptySet := NewStrSet()
		count := 0
		emptySet.Iterator(func(v string) bool {
			count++
			return true
		})

		if count != 0 {
			t.Errorf("Empty set iteration: visited %d items, want 0", count)
		}
	})

	// 测试nil集合
	t.Run("Nil Set", func(t *testing.T) {
		var nilSet *StrSet
		count := 0
		nilSet.Iterator(func(v string) bool {
			count++
			return true
		})

		if count != 0 {
			t.Errorf("Nil set iteration: visited %d items, want 0", count)
		}
	})

	// 测试并发安全性
	t.Run("Concurrent Safety", func(t *testing.T) {
		safeSet := NewStrSet(true)
		for _, v := range testData {
			safeSet.Add(v)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		// 并发迭代
		go func() {
			defer wg.Done()
			safeSet.Iterator(func(v string) bool {
				return true
			})
		}()

		// 并发修改
		go func() {
			defer wg.Done()
			safeSet.Add("new-item")
		}()

		wg.Wait()
	})
}

// TestStrSetConcurrentModification 测试并发修改
func TestStrSetConcurrentModification(t *testing.T) {
	set := NewStrSet(true) // 使用并发安全的集合
	done := make(chan bool)

	// 并发添加和删除
	go func() {
		for i := 0; i < 100; i++ {
			set.Add(strconv.Itoa(i))
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			set.Remove(strconv.Itoa(i))
		}
		done <- true
	}()

	// 等待操作完成
	<-done
	<-done

	// 验证集合状态
	if size := set.Size(); size > 100 {
		t.Errorf("After concurrent modifications, size = %d, should be <= 100", size)
	}
}

// BenchmarkStrSetOperations 性能测试
func BenchmarkStrSetOperations(b *testing.B) {
	// 测试添加性能
	b.Run("Add", func(b *testing.B) {
		set := NewStrSet()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			set.Add("item" + strconv.Itoa(i))
		}
	})

	// 测试查找性能
	b.Run("Contains", func(b *testing.B) {
		set := NewStrSet()
		for i := 0; i < 1000; i++ {
			set.Add("item" + strconv.Itoa(i))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			set.Contains("item" + strconv.Itoa(i%1000))
		}
	})

	// 测试并集操作性能
	b.Run("Union", func(b *testing.B) {
		set1 := NewStrSetFrom([]string{"a", "b", "c"})
		set2 := NewStrSetFrom([]string{"c", "d", "e"})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			set1.Union(set2)
		}
	})

	// 测试序列化性能
	b.Run("Marshal", func(b *testing.B) {
		set := NewStrSet()
		for i := 0; i < 100; i++ {
			set.Add("item" + strconv.Itoa(i))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			set.MarshalJSON()
		}
	})
}

// TestStrSetMemoryUsage 测试内存使用
func TestStrSetMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory usage test in short mode")
	}

	set := NewStrSet()
	n := 100000

	// 添加大量数据
	for i := 0; i < n; i++ {
		set.Add("item" + strconv.Itoa(i))
	}

	if set.Size() != n {
		t.Errorf("Size = %d, want %d", set.Size(), n)
	}

	// 清理并验证内存释放
	set.Clear()
	if set.Size() != 0 {
		t.Error("After Clear, size should be 0")
	}
}

// BenchmarkStrSetIterator 性能测试
func BenchmarkStrSetIterator(b *testing.B) {
	// 准备大数据集
	set := NewStrSet()
	for i := 0; i < 1000; i++ {
		set.Add(strconv.Itoa(i))
	}

	b.ResetTimer()

	// 测试迭代性能
	b.Run("Iterator Performance", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			set.Iterator(func(v string) bool {
				return true
			})
		}
	})

	// 测试提前终止的性能
	b.Run("Early Termination Performance", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			count := 0
			set.Iterator(func(v string) bool {
				count++
				return count < 10 // 只遍历前10个元素
			})
		}
	})
}
