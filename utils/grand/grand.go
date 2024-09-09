// Package grand 提供高性能的随机字节/数字/字符串生成功能。
package grand

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"sync"
	"time"
)

const (
	// bufferSize 是每个缓冲区的大小
	bufferSize = 1024
	// chanSize 是缓冲通道的大小
	chanSize = 1000
)

var (
	// bufferChan 存储随机字节的缓冲通道
	bufferChan = make(chan []byte, chanSize)
	// bufferPool 用于重用缓冲区
	bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, bufferSize)
		},
	}
	letters      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // 52
	symbols      = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"                   // 32
	digits       = "0123456789"                                           // 10
	characters   = letters + digits + symbols                             // 94
	stopChan     = make(chan struct{})
	stopped      bool
	globalCtx    context.Context
	globalCancel context.CancelFunc
	// 用于确保 init 只被调用一次
	initOnce sync.Once
)

// init 初始化随机数生成器
func init() {
	initOnce.Do(func() {
		globalCtx, globalCancel = context.WithCancel(context.Background())
		go generateRandomBytes(globalCtx)
	})
}
func Stop() {
	if !stopped {
		if globalCancel != nil {
			globalCancel()
		}
		close(stopChan)
		// 清空 bufferChan
		for len(bufferChan) > 0 {
			<-bufferChan
		}
		stopped = true
	}
}

func generateRandomBytes(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-stopChan:
			return
		default:
			buffer := bufferPool.Get().([]byte)
			n, err := rand.Read(buffer)
			if err != nil {
				// 使用错误日志记录而不是panic
				// log.Error("从系统读取随机缓冲区时出错:", err)
				time.Sleep(time.Millisecond * 100)
				continue
			}

			for i := 0; i <= n-4; i += 4 {
				select {
				case bufferChan <- buffer[i : i+4]:
				case <-stopChan:
					return
				}
			}

			bufferPool.Put(buffer)
		}
	}
}

// Intn 返回一个介于0和max之间的随机整数: [0, max)
func Intn(max int) int {
	if max <= 0 {
		return max
	}
	n := int(binary.LittleEndian.Uint32(<-bufferChan)) % max
	if (max > 0 && n < 0) || (max < 0 && n > 0) {
		return -n
	}
	return n
}

// B 返回指定长度的随机字节切片
func B(n int) []byte {
	if n <= 0 {
		return nil
	}
	i := 0
	b := make([]byte, n)
	for {
		copy(b[i:], <-bufferChan)
		i += 4
		if i >= n {
			break
		}
	}
	return b
}

// N 返回一个介于min和max之间的随机整数: [min, max]
func N(min, max int) int {
	if min >= max {
		return min
	}
	if min >= 0 {
		return Intn(max-min+1) + min
	}
	return Intn(max+(0-min)+1) - (0 - min)
}

// S 返回一个包含数字和字母的随机字符串，长度为n
func S(n int, symbols ...bool) string {
	if n <= 0 {
		return ""
	}
	var (
		b           = make([]byte, n)
		numberBytes = B(n)
	)
	for i := range b {
		if len(symbols) > 0 && symbols[0] {
			b[i] = characters[numberBytes[i]%94]
		} else {
			b[i] = characters[numberBytes[i]%62]
		}
	}
	return string(b)
}

// D 返回一个介于min和max之间的随机时间间隔
func D(min, max time.Duration) time.Duration {
	multiple := int64(1)
	if min != 0 {
		for min%10 == 0 {
			multiple *= 10
			min /= 10
			max /= 10
		}
	}
	n := int64(N(int(min), int(max)))
	return time.Duration(n * multiple)
}

// Str 从给定的字符串s中随机选择n个字符
func Str(s string, n int) string {
	if n <= 0 {
		return ""
	}
	var (
		b     = make([]rune, n)
		runes = []rune(s)
	)
	if len(runes) <= 255 {
		numberBytes := B(n)
		for i := range b {
			b[i] = runes[int(numberBytes[i])%len(runes)]
		}
	} else {
		for i := range b {
			b[i] = runes[Intn(len(runes))]
		}
	}
	return string(b)
}

// Digits 返回一个只包含数字的随机字符串，长度为n
func Digits(n int) string {
	if n <= 0 {
		return ""
	}
	var (
		b           = make([]byte, n)
		numberBytes = B(n)
	)
	for i := range b {
		b[i] = digits[numberBytes[i]%10]
	}
	return string(b)
}

// Letters 返回一个只包含字母的随机字符串，长度为n
func Letters(n int) string {
	if n <= 0 {
		return ""
	}
	var (
		b           = make([]byte, n)
		numberBytes = B(n)
	)
	for i := range b {
		b[i] = letters[numberBytes[i]%52]
	}
	return string(b)
}

// Symbols 返回一个只包含符号的随机字符串，长度为n
func Symbols(n int) string {
	if n <= 0 {
		return ""
	}
	var (
		b           = make([]byte, n)
		numberBytes = B(n)
	)
	for i := range b {
		b[i] = symbols[numberBytes[i]%32]
	}
	return string(b)
}

// Perm 返回[0,n)的随机排列
func Perm(n int) []int {
	m := make([]int, n)
	for i := 0; i < n; i++ {
		j := Intn(i + 1)
		m[i] = m[j]
		m[j] = i
	}
	return m
}

// Meet 计算给定概率num/total是否满足
func Meet(num, total int) bool {
	return Intn(total) < num
}

// MeetProb 计算给定概率是否满足
func MeetProb(prob float32) bool {
	return Intn(1e7) < int(prob*1e7)
}
