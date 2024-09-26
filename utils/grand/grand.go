// Package grand 提供高性能的随机字节/数字/字符串生成功能。
package grand

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"log"
	"sync"
	"sync/atomic"
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
	bufferChan chan []byte
	// bufferPool 用于重用缓冲区
	bufferPool sync.Pool
	letters    = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") // 52
	symbols    = []byte("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~")                   // 32
	digits     = []byte("0123456789")                                           // 10
	characters = append(append(letters, digits...), symbols...)                 // 94
	// isRunning 用原子操作来保证并发安全
	isRunning    int32
	globalCtx    context.Context
	globalCancel context.CancelFunc
	// 用于确保 init 只被调用一次
	initOnce sync.Once
)

// init 初始化随机数生成器
func init() {
	initOnce.Do(func() {
		bufferChan = make(chan []byte, chanSize)
		bufferPool = sync.Pool{
			New: func() interface{} {
				return make([]byte, bufferSize)
			},
		}
		globalCtx, globalCancel = context.WithCancel(context.Background())
		atomic.StoreInt32(&isRunning, 1)
		go generateRandomBytes(globalCtx)
	})
}

// Stop 停止随机数生成器
func Stop() {
	if atomic.CompareAndSwapInt32(&isRunning, 1, 0) {
		globalCancel()
		// 清空 bufferChan
		for len(bufferChan) > 0 {
			<-bufferChan
		}
		// 等待一小段时间，确保所有正在进行的操作都已完成
		time.Sleep(time.Millisecond * 10)
	}
}

func generateRandomBytes(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in generateRandomBytes: %v", r)
			// 重启 goroutine
			go generateRandomBytes(ctx)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			buffer := bufferPool.Get().([]byte)
			n, err := rand.Read(buffer)
			if err != nil {
				log.Printf("从系统读取随机缓冲区时出错: %v", err)
				time.Sleep(time.Millisecond * 100)
				continue
			}

			for i := 0; i <= n-4; i += 4 {
				select {
				case bufferChan <- buffer[i : i+4]:
				case <-ctx.Done():
					bufferPool.Put(buffer)
					return
				default:
					// 如果 channel 已满，丢弃这个 buffer
					break
				}
			}

			bufferPool.Put(buffer)
		}
	}
}

// getRandomBytes 从 bufferChan 中获取随机字节
func getRandomBytes() ([]byte, error) {
	select {
	case b := <-bufferChan:
		return b, nil
	case <-time.After(time.Second):
		return nil, errors.New("获取随机字节超时")
	}
}

// Intn 返回一个介于0和max之间的随机整数: [0, max)
func Intn(max int) int {
	if atomic.LoadInt32(&isRunning) == 0 || max <= 0 {
		return 0
	}
	b, err := getRandomBytes()
	if err != nil {
		log.Printf("获取随机字节失败: %v", err)
		return 0
	}
	n := int(binary.LittleEndian.Uint32(b)) % max
	if n < 0 {
		n = -n
	}
	return n
}

// B 返回指定长度的随机字节切片
func B(n int) []byte {
	if n <= 0 {
		return nil
	}
	b := make([]byte, n)
	for i := 0; i < n; i += 4 {
		randomBytes, err := getRandomBytes()
		if err != nil {
			log.Printf("获取随机字节失败: %v", err)
			continue
		}
		copy(b[i:], randomBytes[:min(4, n-i)])
	}
	return b
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// N 返回一个介于min和max之间的随机整数: [min, max]
func N(min, max int) int {
	if min >= max {
		return min
	}
	return Intn(max-min+1) + min
}

// S 返回一个包含数字和字母的随机字符串，长度为n
func S(n int, symbols ...bool) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	var src []byte
	if len(symbols) > 0 && symbols[0] {
		src = characters
	} else {
		src = letters
	}
	for i := range b {
		b[i] = src[Intn(len(src))]
	}
	return string(b)
}

// D 返回一个介于min和max之间的随机时间间隔
func D(min, max time.Duration) time.Duration {
	n := int64(N(int(min), int(max)))
	return time.Duration(n)
}

// Str 从给定的字符串s中随机选择n个字符
func Str(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[Intn(len(runes))]
	}
	return string(b)
}

// Digits 返回一个只包含数字的随机字符串，长度为n
func Digits(n int) string {
	return randomString(n, digits)
}

// Letters 返回一个只包含字母的随机字符串，长度为n
func Letters(n int) string {
	return randomString(n, letters)
}

// Symbols 返回一个只包含符号的随机字符串，长度为n
func Symbols(n int) string {
	return randomString(n, symbols)
}

// randomString 是一个通用的随机字符串生成函数
func randomString(n int, charset []byte) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[Intn(len(charset))]
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
