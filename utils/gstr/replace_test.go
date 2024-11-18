package gstr

import (
	"fmt"
	"strings"
	"testing"
)

// 测试辅助函数：检查结果是否与预期相符
func assertEqual(t *testing.T, got, want string, testName string) {
	t.Helper()
	if got != want {
		t.Errorf("%s failed: got %q, want %q", testName, got, want)
	}
}

// Replace函数测试用例
func TestReplace(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		search   string
		replace  string
		count    []int
		expected string
	}{
		{
			name:     "基本替换测试",
			origin:   "hello world hello",
			search:   "hello",
			replace:  "hi",
			expected: "hi world hi",
		},
		{
			name:     "替换一次",
			origin:   "hello world hello",
			search:   "hello",
			replace:  "hi",
			count:    []int{1},
			expected: "hi world hello",
		},
		{
			name:     "空字符串测试",
			origin:   "",
			search:   "hello",
			replace:  "hi",
			expected: "",
		},
		{
			name:     "搜索串为空",
			origin:   "hello",
			search:   "",
			replace:  "hi",
			expected: "hello", // 空搜索串时应该返回原字符串
		},
		{
			name:     "替换串为空",
			origin:   "hello world",
			search:   "hello",
			replace:  "",
			expected: " world",
		},
		{
			name:     "特殊字符测试",
			origin:   "hello#world@hello",
			search:   "#",
			replace:  "@",
			expected: "hello@world@hello",
		},
		{
			name:     "全部替换为空",
			origin:   "hello",
			search:   "hello",
			replace:  "",
			expected: "",
		},
		{
			name:     "连续替换测试",
			origin:   "hellohellohello",
			search:   "hello",
			replace:  "hi",
			expected: "hihihi",
		},
		{
			name:     "部分匹配测试",
			origin:   "hello help hell",
			search:   "hel",
			replace:  "hal",
			expected: "hallo halp hall",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Replace(tt.origin, tt.search, tt.replace, tt.count...)
			if got != tt.expected {
				t.Errorf("%s failed: got %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}

// ReplaceI函数测试用例
func TestReplaceI(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		search   string
		replace  string
		count    []int
		expected string
	}{
		{
			name:     "大小写不敏感替换",
			origin:   "Hello World HELLO",
			search:   "hello",
			replace:  "hi",
			expected: "hi World hi",
		},
		{
			name:     "混合大小写替换",
			origin:   "HeLLo WoRLD hEllO",
			search:   "HELLO",
			replace:  "hi",
			expected: "hi WoRLD hi",
		},
		{
			name:     "替换一次-大小写不敏感",
			origin:   "Hello World HELLO",
			search:   "hello",
			replace:  "hi",
			count:    []int{1},
			expected: "hi World HELLO",
		},
		{
			name:     "空字符串测试",
			origin:   "",
			search:   "hello",
			replace:  "hi",
			expected: "",
		},
		{
			name:     "Unicode字符测试",
			origin:   "你好Hello你好",
			search:   "你好",
			replace:  "哈喽",
			expected: "哈喽Hello哈喽",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceI(tt.origin, tt.search, tt.replace, tt.count...)
			assertEqual(t, got, tt.expected, tt.name)
		})
	}
}

// ReplaceByArray函数测试用例
func TestReplaceByArray(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		array    []string
		expected string
	}{
		{
			name:     "基本数组替换测试",
			origin:   "hello world hello",
			array:    []string{"hello", "hi", "world", "earth"},
			expected: "hi earth hi",
		},
		{
			name:     "空数组测试",
			origin:   "hello world",
			array:    []string{},
			expected: "hello world",
		},
		{
			name:     "奇数长度数组测试",
			origin:   "hello world",
			array:    []string{"hello", "hi", "world"},
			expected: "hi world",
		},
		{
			name:     "包含空字符串的数组测试",
			origin:   "hello world",
			array:    []string{"hello", "", "world", "earth"},
			expected: " earth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceByArray(tt.origin, tt.array)
			assertEqual(t, got, tt.expected, tt.name)
		})
	}
}

// ReplaceIByArray函数测试用例
func TestReplaceIByArray(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		array    []string
		expected string
	}{
		{
			name:     "大小写不敏感数组替换",
			origin:   "Hello World HELLO",
			array:    []string{"hello", "hi", "WORLD", "earth"},
			expected: "hi earth hi",
		},
		{
			name:     "空数组测试",
			origin:   "Hello World",
			array:    []string{},
			expected: "Hello World",
		},
		{
			name:     "混合大小写测试",
			origin:   "HeLLo WoRLD",
			array:    []string{"hello", "hi", "world", "earth"},
			expected: "hi earth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceIByArray(tt.origin, tt.array)
			assertEqual(t, got, tt.expected, tt.name)
		})
	}
}

// ReplaceByMap函数测试用例
func TestReplaceByMap(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		replaces map[string]string
		expected string
	}{
		{
			name:   "基本Map替换测试",
			origin: "hello world hello",
			replaces: map[string]string{
				"hello": "hi",
				"world": "earth",
			},
			expected: "hi earth hi",
		},
		{
			name:     "空Map测试",
			origin:   "hello world",
			replaces: map[string]string{},
			expected: "hello world",
		},
		{
			name:   "特殊字符Map测试",
			origin: "hello#world@hello",
			replaces: map[string]string{
				"#":     "@",
				"hello": "hi",
			},
			expected: "hi@world@hi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceByMap(tt.origin, tt.replaces)
			assertEqual(t, got, tt.expected, tt.name)
		})
	}
}

// ReplaceIByMap函数测试用例
func TestReplaceIByMap(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		replaces map[string]string
		expected string
	}{
		{
			name:   "大小写不敏感Map替换",
			origin: "Hello World HELLO",
			replaces: map[string]string{
				"hello": "hi",
				"world": "earth",
			},
			expected: "hi earth hi",
		},
		{
			name:     "空Map测试",
			origin:   "Hello World",
			replaces: map[string]string{},
			expected: "Hello World",
		},
		{
			name:   "混合大小写测试",
			origin: "HeLLo WoRLD",
			replaces: map[string]string{
				"HELLO": "hi",
				"world": "earth",
			},
			expected: "hi earth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceIByMap(tt.origin, tt.replaces)
			assertEqual(t, got, tt.expected, tt.name)
		})
	}
}

// 性能测试
func BenchmarkReplace(b *testing.B) {
	longString := strings.Repeat("hello world ", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Replace(longString, "hello", "hi")
	}
}

func BenchmarkReplaceI(b *testing.B) {
	longString := strings.Repeat("Hello World ", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReplaceI(longString, "hello", "hi")
	}
}

func BenchmarkReplaceByArray(b *testing.B) {
	longString := strings.Repeat("hello world ", 100)
	array := []string{"hello", "hi", "world", "earth"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReplaceByArray(longString, array)
	}
}

func BenchmarkReplaceByMap(b *testing.B) {
	longString := strings.Repeat("hello world ", 100)
	replaces := map[string]string{
		"hello": "hi",
		"world": "earth",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReplaceByMap(longString, replaces)
	}
}

// 并发测试
func TestConcurrentUsage(t *testing.T) {
	// 测试并发访问是否安全
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(val int) {
			str := fmt.Sprintf("hello%d world%d", val, val)
			result := ReplaceI(str, "hello", "hi")
			if !strings.Contains(result, "hi") {
				t.Errorf("Concurrent test failed for input: %s", str)
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 100; i++ {
		<-done
	}
}
