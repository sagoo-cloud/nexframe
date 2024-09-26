package grand

import (
	"context"
	"math"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestB(t *testing.T) {
	lengths := []int{10, 100, 1000}
	for _, length := range lengths {
		b := B(length)
		if len(b) != length {
			t.Errorf("B(%d) 返回的切片长度为 %d", length, len(b))
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
				t.Errorf("N(%d, %d) 返回 %d，超出范围", tc.min, tc.max, n)
			}
		}
	}
}

func TestS(t *testing.T) {
	lengths := []int{10, 100, 1000}
	for _, length := range lengths {
		s := S(length)
		t.Log(s)
		if len(s) != length {
			t.Errorf("S(%d) 返回字符串长度为 %d", length, len(s))
		}
		if !isAlphanumeric(s) {
			t.Errorf("S(%d) 返回非字母数字字符串: %s", length, s)
		}
	}

	// 测试包含符号的情况
	s := S(100, true)
	if !containsSymbols(s) {
		t.Errorf("S(100, true) 没有返回包含符号的字符串: %s", s)
	}
}

func TestD(t *testing.T) {
	min := time.Second
	max := time.Minute
	for i := 0; i < 1000; i++ {
		d := D(min, max)
		if d < min || d > max {
			t.Errorf("D(%v, %v) 返回 %v，超出范围", min, max, d)
		}
	}
}

func TestStr(t *testing.T) {
	source := "abcdefghijklmnopqrstuvwxyz"
	length := 100
	s := Str(source, length)
	if len(s) != length {
		t.Errorf("Str() 返回字符串长度为 %d，期望 %d", len(s), length)
	}
	for _, char := range s {
		if !strings.ContainsRune(source, char) {
			t.Errorf("Str() 返回的字符串包含意外字符: %c", char)
		}
	}
}

func TestDigits(t *testing.T) {
	length := 100
	s := Digits(length)
	if len(s) != length {
		t.Errorf("Digits(%d) 返回字符串长度为 %d", length, len(s))
	}
	if !isNumeric(s) {
		t.Errorf("Digits(%d) 返回非数字字符串: %s", length, s)
	}
}

func TestLetters(t *testing.T) {
	length := 100
	s := Letters(length)
	if len(s) != length {
		t.Errorf("Letters(%d) 返回字符串长度为 %d", length, len(s))
	}
	if !isAlphabetic(s) {
		t.Errorf("Letters(%d) 返回非字母字符串: %s", length, s)
	}
}

func TestSymbols(t *testing.T) {
	length := 100
	s := Symbols(length)
	if len(s) != length {
		t.Errorf("Symbols(%d) 返回字符串长度为 %d", length, len(s))
	}
	if !isSymbolic(s) {
		t.Errorf("Symbols(%d) 返回不包含符号的字符串: %s", length, s)
	}
}

func TestPerm(t *testing.T) {
	n := 100
	p := Perm(n)
	if len(p) != n {
		t.Errorf("Perm(%d) 返回切片长度为 %d", n, len(p))
	}
	seen := make(map[int]bool)
	for _, v := range p {
		if v < 0 || v >= n {
			t.Errorf("Perm(%d) 返回超出范围的值: %d", n, v)
		}
		if seen[v] {
			t.Errorf("Perm(%d) 返回重复值: %d", n, v)
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
		t.Errorf("Meet(1, 2) 概率为 %f，期望接近 %f", actual, threshold)
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
		t.Errorf("MeetProb(0.3) 概率为 %f，期望接近 %f", actual, threshold)
	}
}

func TestStop(t *testing.T) {
	// 重置状态
	atomic.StoreInt32(&isRunning, 1)
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

	// 检查 isRunning 是否已经被设置为 0
	if atomic.LoadInt32(&isRunning) != 0 {
		t.Error("Stop() 没有将 isRunning 设置为 0")
	}

	// 尝试再次生成随机数，应该会返回 0
	n := Intn(100)
	if n != 0 {
		t.Errorf("停止后 Intn(100) 返回 %d，期望返回 0", n)
	}

	// 重新初始化，以便其他测试可以正常运行
	atomic.StoreInt32(&isRunning, 1)
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
	return strings.ContainsAny(s, string(symbols))
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
		if !strings.ContainsRune(string(symbols), char) {
			return false
		}
	}
	return true
}
