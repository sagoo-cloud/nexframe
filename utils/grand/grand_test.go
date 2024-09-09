package grand

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"
)

func TestB(t *testing.T) {
	lengths := []int{10, 100, 1000}
	for _, length := range lengths {
		b := B(length)
		if len(b) != length {
			t.Errorf("B(%d) returned slice of length %d", length, len(b))
		}
	}
}

func TestN(t *testing.T) {
	testCases := []struct {
		min, max int
	}{
		{0, 10},
		{-10, 10},
		{-100, -50},
	}

	for _, tc := range testCases {
		for i := 0; i < 1000; i++ {
			n := N(tc.min, tc.max)
			if n < tc.min || n > tc.max {
				t.Errorf("N(%d, %d) returned %d, which is out of range", tc.min, tc.max, n)
			}
		}
	}
}

func TestS(t *testing.T) {
	lengths := []int{10, 100, 1000}
	for _, length := range lengths {
		s := S(length)
		if len(s) != length {
			t.Errorf("S(%d) returned string of length %d", length, len(s))
		}
		if !isAlphanumeric(s) {
			t.Errorf("S(%d) returned non-alphanumeric string: %s", length, s)
		}
	}

	// 测试包含符号的情况
	s := S(100, true)
	if !containsSymbols(s) {
		t.Errorf("S(100, true) did not return string with symbols: %s", s)
	}
}

func TestD(t *testing.T) {
	min := time.Second
	max := time.Minute
	for i := 0; i < 1000; i++ {
		d := D(min, max)
		if d < min || d > max {
			t.Errorf("D(%v, %v) returned %v, which is out of range", min, max, d)
		}
	}
}

func TestStr(t *testing.T) {
	source := "abcdefghijklmnopqrstuvwxyz"
	length := 100
	s := Str(source, length)
	if len(s) != length {
		t.Errorf("Str() returned string of length %d, expected %d", len(s), length)
	}
	for _, char := range s {
		if !strings.ContainsRune(source, char) {
			t.Errorf("Str() returned string containing unexpected character: %c", char)
		}
	}
}

func TestDigits(t *testing.T) {
	length := 100
	s := Digits(length)
	if len(s) != length {
		t.Errorf("Digits(%d) returned string of length %d", length, len(s))
	}
	if !isNumeric(s) {
		t.Errorf("Digits(%d) returned non-numeric string: %s", length, s)
	}
}

func TestLetters(t *testing.T) {
	length := 100
	s := Letters(length)
	if len(s) != length {
		t.Errorf("Letters(%d) returned string of length %d", length, len(s))
	}
	if !isAlphabetic(s) {
		t.Errorf("Letters(%d) returned non-alphabetic string: %s", length, s)
	}
}

func TestSymbols(t *testing.T) {
	length := 100
	s := Symbols(length)
	if len(s) != length {
		t.Errorf("Symbols(%d) returned string of length %d", length, len(s))
	}
	if !isSymbolic(s) {
		t.Errorf("Symbols(%d) returned string without symbols: %s", length, s)
	}
}

func TestPerm(t *testing.T) {
	n := 100
	p := Perm(n)
	if len(p) != n {
		t.Errorf("Perm(%d) returned slice of length %d", n, len(p))
	}
	seen := make(map[int]bool)
	for _, v := range p {
		if v < 0 || v >= n {
			t.Errorf("Perm(%d) returned out of range value: %d", n, v)
		}
		if seen[v] {
			t.Errorf("Perm(%d) returned duplicate value: %d", n, v)
		}
		seen[v] = true
	}
}

func TestMeet(t *testing.T) {
	iterations := 10000
	threshold := 0.5
	count := 0
	for i := 0; i < iterations; i++ {
		if Meet(1, 2) {
			count++
		}
	}
	actual := float64(count) / float64(iterations)
	if math.Abs(actual-threshold) > 0.05 {
		t.Errorf("Meet(1, 2) probability is %f, expected around %f", actual, threshold)
	}
}

func TestMeetProb(t *testing.T) {
	iterations := 10000
	threshold := 0.3
	count := 0
	for i := 0; i < iterations; i++ {
		if MeetProb(0.3) {
			count++
		}
	}
	actual := float64(count) / float64(iterations)
	if math.Abs(actual-threshold) > 0.05 {
		t.Errorf("MeetProb(0.3) probability is %f, expected around %f", actual, threshold)
	}
}

func TestStop(t *testing.T) {
	// 重置状态
	stopped = false
	stopChan = make(chan struct{})
	globalCtx, globalCancel = context.WithCancel(context.Background())
	go generateRandomBytes(globalCtx)

	// 执行一些操作以确保 generateRandomBytes 已经开始生成随机数
	for i := 0; i < 100; i++ {
		Intn(100)
	}

	// 调用 Stop 函数
	Stop()

	// 等待一小段时间，让 goroutine 有机会停止
	time.Sleep(time.Millisecond * 100)

	// 检查 globalCtx 是否已经被取消
	select {
	case <-globalCtx.Done():
		// 这是预期的行为
	default:
		t.Error("Stop() did not cancel the global context")
	}

	// 尝试再次生成随机数，应该会失败或阻塞
	done := make(chan bool)
	go func() {
		Intn(100)
		done <- true
	}()

	select {
	case <-done:
		t.Error("Random number generation succeeded after Stop()")
	case <-time.After(time.Millisecond * 500):
		// 这是预期的行为 - 生成随机数应该被阻塞或失败
	}

	// 重新初始化，以便其他测试可以正常运行
	stopped = false
	stopChan = make(chan struct{})
	globalCtx, globalCancel = context.WithCancel(context.Background())
	go generateRandomBytes(globalCtx)
}

// 辅助函数

func isAlphanumeric(s string) bool {
	for _, char := range s {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	return true
}

func containsSymbols(s string) bool {
	return strings.ContainsAny(s, symbols)
}

func isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func isAlphabetic(s string) bool {
	for _, char := range s {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
			return false
		}
	}
	return true
}

func isSymbolic(s string) bool {
	for _, char := range s {
		if !strings.ContainsRune(symbols, char) {
			return false
		}
	}
	return true
}
