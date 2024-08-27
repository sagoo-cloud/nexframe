package gcharset

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSupported(t *testing.T) {
	tests := []struct {
		charset  string
		expected bool
	}{
		{"UTF-8", true},
		{"utf-8", true},
		{"GBK", true},
		{"gbk", true},
		{"GB18030", true},
		{"Big5", true},
		{"EUCJP", true},
		{"EUCKR", true},
		{"ISO-8859-1", true},
		{"NonExistentCharset", false},
	}

	for _, test := range tests {
		t.Run(test.charset, func(t *testing.T) {
			result := Supported(test.charset)
			assert.Equal(t, test.expected, result, "Supported(%s) should return %v", test.charset, test.expected)
		})
	}
}

func TestConvert(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		srcCharset string
		dstCharset string
		expected   string
		expectErr  bool
	}{
		{
			name:       "UTF-8 to GBK",
			src:        "你好，世界",
			srcCharset: "UTF-8",
			dstCharset: "GBK",
			expected:   "\xC4\xE3\xBA\xC3\xA3\xAC\xCA\xC0\xBD\xE7",
			expectErr:  false,
		},
		{
			name:       "GBK to UTF-8",
			src:        "\xC4\xE3\xBA\xC3\xA3\xAC\xCA\xC0\xBD\xE7",
			srcCharset: "GBK",
			dstCharset: "UTF-8",
			expected:   "你好，世界",
			expectErr:  false,
		},
		{
			name:       "UTF-8 to UTF-8",
			src:        "Hello, World!",
			srcCharset: "UTF-8",
			dstCharset: "UTF-8",
			expected:   "Hello, World!",
			expectErr:  false,
		},
		{
			name:       "Invalid source charset",
			src:        "Test",
			srcCharset: "InvalidCharset",
			dstCharset: "UTF-8",
			expected:   "",
			expectErr:  true,
		},
		{
			name:       "Invalid destination charset",
			src:        "Test",
			srcCharset: "UTF-8",
			dstCharset: "InvalidCharset",
			expected:   "",
			expectErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Convert(test.dstCharset, test.srcCharset, test.src)
			if test.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestToUTF8(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		srcCharset string
		expected   string
		expectErr  bool
	}{
		{
			name:       "GBK to UTF-8",
			src:        "\xC4\xE3\xBA\xC3\xA3\xAC\xCA\xC0\xBD\xE7",
			srcCharset: "GBK",
			expected:   "你好，世界",
			expectErr:  false,
		},
		{
			name:       "UTF-8 to UTF-8",
			src:        "Hello, World!",
			srcCharset: "UTF-8",
			expected:   "Hello, World!",
			expectErr:  false,
		},
		{
			name:       "Invalid charset",
			src:        "Test",
			srcCharset: "InvalidCharset",
			expected:   "",
			expectErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ToUTF8(test.srcCharset, test.src)
			if test.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestUTF8To(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		dstCharset string
		expected   string
		expectErr  bool
	}{
		{
			name:       "UTF-8 to GBK",
			src:        "你好，世界",
			dstCharset: "GBK",
			expected:   "\xC4\xE3\xBA\xC3\xA3\xAC\xCA\xC0\xBD\xE7",
			expectErr:  false,
		},
		{
			name:       "UTF-8 to UTF-8",
			src:        "Hello, World!",
			dstCharset: "UTF-8",
			expected:   "Hello, World!",
			expectErr:  false,
		},
		{
			name:       "Invalid charset",
			src:        "Test",
			dstCharset: "InvalidCharset",
			expected:   "",
			expectErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := UTF8To(test.dstCharset, test.src)
			if test.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestConvertWithSizeLimit(t *testing.T) {
	// 创建一个超过大小限制的字符串
	largeString := string(make([]byte, MaxInputSize+1))

	tests := []struct {
		name       string
		src        string
		srcCharset string
		dstCharset string
		expectErr  bool
	}{
		{
			name:       "Normal conversion",
			src:        "Hello, World!",
			srcCharset: "UTF-8",
			dstCharset: "GBK",
			expectErr:  false,
		},
		{
			name:       "Exceed size limit",
			src:        largeString,
			srcCharset: "UTF-8",
			dstCharset: "GBK",
			expectErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ConvertWithSizeLimit(test.dstCharset, test.srcCharset, test.src)
			if test.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
