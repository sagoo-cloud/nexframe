package gstr

import (
	"strings"
	"testing"
)

func TestQuoteMeta(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		chars    []string
		expected string
	}{
		{
			name:     "空字符串测试",
			input:    "",
			expected: "",
		},
		{
			name:     "普通字符串测试",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "默认元字符测试",
			input:    "hello.world*test[abc]",
			expected: `hello\.world\*test\[abc\]`,
		},
		{
			name:     "所有默认元字符测试",
			input:    `.+\($)[^]*?{}`,
			expected: `\.\+\\\(\$\)\[\^\]\*\?\{\}`,
		},
		{
			name:     "自定义字符集测试",
			input:    "hello@world#test",
			chars:    []string{"@#"},
			expected: `hello\@world\#test`,
		},
		{
			name:     "空自定义字符集测试",
			input:    "hello.world",
			chars:    []string{""},
			expected: "hello.world",
		},
		{
			name:     "Unicode字符测试",
			input:    "你好.世界*测试",
			expected: `你好\.世界\*测试`,
		},
		{
			name:     "连续元字符测试",
			input:    "...***{}}",
			expected: `\.\.\.\*\*\*\{\}\}`,
		},
		{
			name:     "混合字符测试",
			input:    "abc123.[]{}*+?",
			expected: `abc123\.\[\]\{\}\*\+\?`,
		},
		{
			name:     "自定义Unicode字符集测试",
			input:    "你好世界",
			chars:    []string{"你好"},
			expected: `\你\好世界`,
		},
		{
			name:     "花括号嵌套测试",
			input:    "{{test}}",
			expected: `\{\{test\}\}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := QuoteMeta(tt.input, tt.chars...)
			if got != tt.expected {
				t.Errorf("QuoteMeta(%q, %v) = %q, want %q",
					tt.input, tt.chars, got, tt.expected)
			}
		})
	}
}

// 性能测试
func BenchmarkQuoteMeta(b *testing.B) {
	testCases := []struct {
		name  string
		input string
		chars []string
	}{
		{
			name:  "短字符串无元字符",
			input: "hello world",
		},
		{
			name:  "短字符串带元字符",
			input: "hello.world*test",
		},
		{
			name:  "长字符串无元字符",
			input: strings.Repeat("hello world ", 100),
		},
		{
			name:  "长字符串带元字符",
			input: strings.Repeat("hello.world*test ", 100),
		},
		{
			name:  "自定义字符集",
			input: "hello@world#test",
			chars: []string{"@#"},
		},
		{
			name:  "包含所有元字符",
			input: `.+\($)[^]*?{}`,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				QuoteMeta(tc.input, tc.chars...)
			}
		})
	}
}

// 并发安全性测试
func TestQuoteMetaConcurrent(t *testing.T) {
	const goroutines = 100
	done := make(chan bool, goroutines)

	inputs := []string{
		"hello.world",
		"test{abc}",
		`.+\($)[^]*?{}`,
		"你好.世界*测试",
	}

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			input := inputs[n%len(inputs)]
			result := QuoteMeta(input)
			if !strings.Contains(result, `\`) {
				t.Errorf("Concurrent test failed for input: %s", input)
			}
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}
