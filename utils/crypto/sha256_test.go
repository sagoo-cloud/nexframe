package crypto

import (
	"encoding/hex"
	"testing"
)

func TestSha256(t *testing.T) {
	tests := []struct {
		input    []byte
		expected string
	}{
		{[]byte("Hello, World!"), "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"},
		{[]byte(""), "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{[]byte("The quick brown fox jumps over the lazy dog"), "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592"},
		{[]byte{0, 1, 2, 3, 4, 5}, "17e88db187afd62c16e5debf3e6527cd006bc012bc90b51a810cd80c2d511f43"},
		{[]byte("一二三四五"), "3c040c8ec536512dd70e2c62397397b1362bf02a42842ac12159c7680a1b9690"},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := Sha256(tt.input)
			if result != tt.expected {
				t.Errorf("Sha256(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSha256String(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello, World!", "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"},
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"The quick brown fox jumps over the lazy dog", "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592"},
		{"一二三四五", "3c040c8ec536512dd70e2c62397397b1362bf02a42842ac12159c7680a1b9690"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Sha256String(tt.input)
			if result != tt.expected {
				t.Errorf("Sha256String(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSha256WithError(t *testing.T) {
	tests := []struct {
		input    []byte
		expected string
	}{
		{[]byte("Hello, World!"), "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"},
		{[]byte(""), "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{[]byte("The quick brown fox jumps over the lazy dog"), "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592"},
		{[]byte{0, 1, 2, 3, 4, 5}, "17e88db187afd62c16e5debf3e6527cd006bc012bc90b51a810cd80c2d511f43"},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result, err := Sha256WithError(tt.input)
			if err != nil {
				t.Errorf("Sha256WithError(%q) returned unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("Sha256WithError(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSha256Consistency(t *testing.T) {
	input := []byte("Test consistency across functions")

	result1 := Sha256(input)
	result2 := Sha256String(string(input))
	result3, _ := Sha256WithError(input)

	if result1 != result2 || result2 != result3 {
		t.Errorf("Inconsistent results: Sha256=%v, Sha256String=%v, Sha256WithError=%v", result1, result2, result3)
	}
}

func TestSha256LargeInput(t *testing.T) {
	largeInput := make([]byte, 1000000) // 1 MB of data
	for i := range largeInput {
		largeInput[i] = byte(i % 256)
	}

	result := Sha256(largeInput)
	if len(result) != 64 { // SHA-256 输出应该总是 64 个字符的十六进制字符串
		t.Errorf("Sha256(largeInput) returned hash of unexpected length: got %d, want 64", len(result))
	}

	_, err := hex.DecodeString(result)
	if err != nil {
		t.Errorf("Sha256(largeInput) returned invalid hex string: %v", err)
	}
}
