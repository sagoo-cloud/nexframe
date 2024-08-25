package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"sync"
)

var (
	// 全局随机数生成器
	globalRand = rand.Reader
	// 用于保护随机数生成的互斥锁
	randMutex sync.Mutex
)

// Rand 返回范围 [min, max] 内的随机整数
func Rand(min, max int) (int, error) {
	if min > max {
		return 0, fmt.Errorf("最小值必须小于或等于最大值")
	}
	randMutex.Lock()
	defer randMutex.Unlock()
	n, err := rand.Int(globalRand, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + min, nil
}

// RandFloat 返回范围 [min, max) 内的随机浮点数
func RandFloat(min, max float64) (float64, error) {
	if min >= max {
		return 0, fmt.Errorf("最小值必须小于最大值")
	}
	randMutex.Lock()
	defer randMutex.Unlock()

	// 生成一个大整数，表示 [0, 2^53) 范围内的随机数
	maxInt := big.NewInt(1 << 53)
	n, err := rand.Int(globalRand, maxInt)
	if err != nil {
		return 0, err
	}

	// 将大整数转换为 [0, 1) 范围内的浮点数
	f := float64(n.Int64()) / float64(maxInt.Int64())

	// 将结果映射到 [min, max) 范围
	return f*(max-min) + min, nil
}

var (
	// 所有可用字符
	allChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 仅字母
	lettersOnly = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// 仅小写字母
	lowercaseOnly = "abcdefghijklmnopqrstuvwxyz"
	// 仅大写字母
	uppercaseOnly = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// 仅数字
	numbersOnly = "0123456789"
)

// randStringFromCharset 从给定的字符集中生成指定长度的随机字符串
func randStringFromCharset(length int, charset string) (string, error) {
	if length < 0 {
		return "", fmt.Errorf("长度必须为非负数")
	}
	randMutex.Lock()
	defer randMutex.Unlock()
	result := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(globalRand, charsetLength)
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}

// RandString 返回指定长度的随机字符串（包括字母和数字）
func RandString(length int) (string, error) {
	return randStringFromCharset(length, allChars)
}

// RandStringNoNumber 返回指定长度的随机字符串（仅包含字母）
func RandStringNoNumber(length int) (string, error) {
	return randStringFromCharset(length, lettersOnly)
}

// RandLowerString 返回指定长度的随机小写字母字符串
func RandLowerString(length int) (string, error) {
	return randStringFromCharset(length, lowercaseOnly)
}

// RandUpperString 返回指定长度的随机大写字母字符串
func RandUpperString(length int) (string, error) {
	return randStringFromCharset(length, uppercaseOnly)
}

// RandNumber 返回指定长度的随机数字字符串
func RandNumber(length int) (string, error) {
	return randStringFromCharset(length, numbersOnly)
}

// RandIntSlice 随机打乱整数切片
func RandIntSlice(slice []int) error {
	for i := len(slice) - 1; i > 0; i-- {
		j, err := Rand(0, i)
		if err != nil {
			return err
		}
		slice[i], slice[j] = slice[j], slice[i]
	}
	return nil
}

// UniqueCodeMap 用于生成唯一码的映射
var UniqueCodeMap = map[string][]string{
	"0": {"A", "C"},
	"1": {"D", "E"},
	"2": {"F", "G", "Y"},
	"3": {"H", "I"},
	"4": {"Z", "K", "X"},
	"5": {"L", "M", "W"},
	"6": {"N", "O"},
	"7": {"P", "Q", "Z"},
	"8": {"R", "S", "V"},
	"9": {"T", "U"},
}

// UniqueCode 根据给定的 ID 和最小长度生成唯一码
func UniqueCode(id int64, minLen int) (string, error) {
	ids := fmt.Sprintf("%d", id)
	result := strings.Builder{}

	for _, digit := range ids {
		choices := UniqueCodeMap[string(digit)]
		if len(choices) == 0 {
			return "", fmt.Errorf("ID 中包含无效数字: %c", digit)
		}
		index, err := Rand(0, len(choices)-1)
		if err != nil {
			return "", err
		}
		result.WriteString(choices[index])
	}

	if result.Len() < minLen {
		result.WriteString("B")
		for i := result.Len(); i < minLen; i++ {
			digit, err := RandNumber(1)
			if err != nil {
				return "", err
			}
			choices := UniqueCodeMap[digit]
			index, err := Rand(0, len(choices)-1)
			if err != nil {
				return "", err
			}
			result.WriteString(choices[index])
		}
	}

	return result.String(), nil
}
