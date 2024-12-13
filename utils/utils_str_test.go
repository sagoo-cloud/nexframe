package utils

import (
	"reflect"
	"testing"
)

// TestIsLetterUpper 测试大写字母检查函数
func TestIsLetterUpper(t *testing.T) {
	tests := []struct {
		name string
		b    byte
		want bool
	}{
		{"大写字母A", 'A', true},
		{"大写字母Z", 'Z', true},
		{"小写字母a", 'a', false},
		{"数字", '1', false},
		{"特殊字符", '*', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLetterUpper(tt.b); got != tt.want {
				t.Errorf("IsLetterUpper() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsLetterLower 测试小写字母检查函数
func TestIsLetterLower(t *testing.T) {
	tests := []struct {
		name string
		b    byte
		want bool
	}{
		{"小写字母a", 'a', true},
		{"小写字母z", 'z', true},
		{"大写字母A", 'A', false},
		{"数字", '1', false},
		{"特殊字符", '*', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLetterLower(tt.b); got != tt.want {
				t.Errorf("IsLetterLower() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsLetter 测试字母检查函数
func TestIsLetter(t *testing.T) {
	tests := []struct {
		name string
		b    byte
		want bool
	}{
		{"小写字母a", 'a', true},
		{"大写字母Z", 'Z', true},
		{"数字", '1', false},
		{"特殊字符", '*', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLetter(tt.b); got != tt.want {
				t.Errorf("IsLetter() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsNumeric 测试数字字符串检查函数
func TestIsNumeric(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"正整数", "123", true},
		{"负整数", "-123", true},
		{"正小数", "123.456", true},
		{"负小数", "-123.456", true},
		{"非法小数点位置", "123.", false},
		{"多个小数点", "123.456.789", false},
		{"非数字字符", "123abc", false},
		{"空字符串", "", false},
		{"只有小数点", ".", false},
		{"特殊字符", "12#34", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNumeric(tt.s); got != tt.want {
				t.Errorf("IsNumeric() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestUcFirst 测试首字母大写函数
func TestUcFirst(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"小写开头", "hello", "Hello"},
		{"大写开头", "Hello", "Hello"},
		{"数字开头", "123abc", "123abc"},
		{"空字符串", "", ""},
		{"特殊字符开头", "*abc", "*abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UcFirst(tt.s); got != tt.want {
				t.Errorf("UcFirst() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestReplaceByMap 测试字符串替换函数
func TestReplaceByMap(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		replaces map[string]string
		want     string
	}{
		{
			"基本替换",
			"hello world",
			map[string]string{"hello": "hi", "world": "golang"},
			"hi golang",
		},
		{
			"空字符串",
			"",
			map[string]string{"hello": "hi"},
			"",
		},
		{
			"空映射",
			"hello",
			map[string]string{},
			"hello",
		},
		{
			"重叠替换",
			"aabbcc",
			map[string]string{"aa": "11", "bb": "22", "cc": "33"},
			"112233",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceByMap(tt.origin, tt.replaces); got != tt.want {
				t.Errorf("ReplaceByMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRemoveSymbols 测试符号移除函数
func TestRemoveSymbols(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"混合字符", "Hello123!@#", "Hello123"},
		{"纯符号", "!@#$%", ""},
		{"空字符串", "", ""},
		{"中文字符", "你好123ABC", "你好123ABC"},
		{"特殊符号", "Hello-_World", "HelloWorld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveSymbols(tt.s); got != tt.want {
				t.Errorf("RemoveSymbols() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestEqualFoldWithoutChars 测试忽略大小写和特殊字符的字符串比较函数
func TestEqualFoldWithoutChars(t *testing.T) {
	tests := []struct {
		name string
		s1   string
		s2   string
		want bool
	}{
		{"完全相同", "hello", "hello", true},
		{"大小写不同", "Hello", "hello", true},
		{"带符号", "hello-world", "helloworld", true},
		{"不同字符串", "hello", "world", false},
		{"空字符串", "", "", true},
		{"中文字符", "你好123", "你好123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EqualFoldWithoutChars(tt.s1, tt.s2); got != tt.want {
				t.Errorf("EqualFoldWithoutChars() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSplitAndTrim 测试字符串分割和修剪函数
func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		name          string
		str           string
		delimiter     string
		characterMask []string
		want          []string
	}{
		{
			"基本分割",
			"a,b,c",
			",",
			nil,
			[]string{"a", "b", "c"},
		},
		{
			"带空格",
			" a , b , c ",
			",",
			nil,
			[]string{"a", "b", "c"},
		},
		{
			"自定义修剪字符",
			"*a*,*b*,*c*",
			",",
			[]string{"*"},
			[]string{"a", "b", "c"},
		},
		{
			"空字符串",
			"",
			",",
			nil,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitAndTrim(tt.str, tt.delimiter, tt.characterMask...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitAndTrim() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTrim 测试字符串修剪函数
func TestTrim(t *testing.T) {
	tests := []struct {
		name          string
		str           string
		characterMask []string
		want          string
	}{
		{"默认修剪", " \tHello\n ", nil, "Hello"},
		{"自定义修剪", "*Hello*", []string{"*"}, "Hello"},
		{"空字符串", "", nil, ""},
		{"全是修剪字符", " \t\n", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Trim(tt.str, tt.characterMask...); got != tt.want {
				t.Errorf("Trim() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFormatCmdKey 测试命令键格式化函数
func TestFormatCmdKey(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"基本转换", "MY_CMD_KEY", "my.cmd.key"},
		{"已是小写", "my_cmd_key", "my.cmd.key"},
		{"空字符串", "", ""},
		{"无下划线", "MYCMDKEY", "mycmdkey"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatCmdKey(tt.s); got != tt.want {
				t.Errorf("FormatCmdKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFormatEnvKey 测试环境变量键格式化函数
func TestFormatEnvKey(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"基本转换", "my.env.key", "MY_ENV_KEY"},
		{"已是大写", "MY.ENV.KEY", "MY_ENV_KEY"},
		{"空字符串", "", ""},
		{"无点号", "myenvkey", "MYENVKEY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatEnvKey(tt.s); got != tt.want {
				t.Errorf("FormatEnvKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestStripSlashes 测试反斜杠去除函数
func TestStripSlashes(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{"基本转义", `Hello\\World`, "HelloWorld"},
		{"无转义字符", "HelloWorld", "HelloWorld"},
		{"空字符串", "", ""},
		{"多个转义", `\\Hello\\World\\`, "HelloWorld"},
		{"单个反斜杠", `\Hello`, "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripSlashes(tt.str); got != tt.want {
				t.Errorf("StripSlashes() = %v, want %v", got, tt.want)
			}
		})
	}
}
