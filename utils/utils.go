package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// GetLocalIP 获取服务器内网IP
func GetLocalIP() (string, error) {
	var localIP string
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	// 遍历所有网卡
	for _, i := range interfaces {
		if i.Flags&net.FlagUp == 0 {
			continue // 网卡未开启
		}
		if i.Flags&net.FlagLoopback != 0 {
			continue // 网卡为loopback地址
		}
		addrs, err := i.Addrs()
		if err != nil {
			return "", err
		}
		// 遍历网卡上的所有地址
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // 不是ipv4地址
			}
			localIP = ip.String()
			return localIP, nil
		}
	}
	return "", fmt.Errorf("no local IP address found")
}

// GetPublicIP 获取公网IP
func GetPublicIP() (ip string, err error) {
	resp, err := http.Get("https://ifconfig.co/ip")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			fmt.Println(err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	ip = string(body)
	// 去除空格
	ip = strings.Replace(ip, " ", "", -1)
	// 去除换行符
	ip = strings.Replace(ip, "\n", "", -1)

	return
}

func RemoveRepeatedElementAndEmpty(arr []int) []int {
	newArr := make([]int, 0)
	for _, item := range arr {
		repeat := false
		if len(newArr) > 0 {
			for _, v := range newArr {
				if v == item {
					repeat = true
					break
				}
			}
		}
		if repeat {
			continue
		}
		newArr = append(newArr, item)
	}
	return newArr
}

// RemoveDuplicationMap 数组去重
func RemoveDuplicationMap(arr []string) []string {
	set := make(map[string]struct{}, len(arr))
	j := 0
	for _, v := range arr {
		_, ok := set[v]
		if ok {
			continue
		}
		set[v] = struct{}{}
		arr[j] = v
		j++
	}
	return arr[:j]
}

// Decimal 保留两位小数
func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

// InArray 判断字符串是否存在数组中
func InArray(target string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)
	if index < len(strArray) && strArray[index] == target {
		return true
	}
	return false
}

// FileSize 字节的单位转换 保留两位小数
func FileSize(fileSize int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "EB"}
	var size = float64(fileSize)
	var i int
	for i = 0; size > 1024; i++ {
		size /= 1024
	}
	return fmt.Sprintf("%.2f %s", size, units[i])
}

type fileInfo struct {
	name string
	size int64
}

// WalkDir 获取目录下文件的名称和大小
func WalkDir(dirname string) ([]fileInfo, error) {
	var fileInfos []fileInfo
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fileInfos = append(fileInfos, fileInfo{name: path, size: info.Size()})
		}
		return nil
	})

	return fileInfos, err
}

// DirSize 获取目录下所有文件大小
func DirSize(dirname string) string {
	var (
		s        int64
		files, _ = WalkDir(dirname)
	)
	for _, n := range files {
		s += n.size
	}
	return FileSize(s)
}

func ConvertToStringSlice(data []interface{}) []string {
	result := make([]string, len(data))
	for i, v := range data {
		str, ok := v.(string)
		if !ok {
			// 如果类型断言失败，可以在此处进行相应的错误处理
			// 这里简单地将该元素转换为空字符串
			str = ""
		}
		result[i] = str
	}
	return result
}

// 删除文件
func DeleteFile(name string) error {
	//判断文件是否存在
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	//打开文件
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	//删除文件
	err = os.Remove(name)
	if err != nil {
		return err
	}

	return nil
}

const (
	kilobyte = 1024
	megabyte = 1024 * kilobyte
	gigabyte = 1024 * megabyte
	terabyte = 1024 * gigabyte
	petabyte = 1024 * terabyte
)

// FormatSize 格式化文件大小。
func FormatSize(size int64) string {
	switch {
	case size < kilobyte:
		return strconv.Itoa(int(size)) + "B"
	case size < megabyte:
		return fmt.Sprintf("%.2fK", float64(size)/kilobyte)
	case size < gigabyte:
		return fmt.Sprintf("%.2fM", float64(size)/megabyte)
	case size < terabyte:
		return fmt.Sprintf("%.2fG", float64(size)/gigabyte)
	case size < petabyte:
		return fmt.Sprintf("%.2fT", float64(size)/terabyte)
	default:
		return fmt.Sprintf("%.2fP", float64(size)/petabyte)
	}
}

func ValidatePassword(password string, minimumLength int, requireComplexity int, requireDigit int, requireLowercase int, requireUppercase int) (err error) {
	//判断密码长度
	if len(password) < minimumLength {
		err = fmt.Errorf("密码长度必须大于和等于%d位", minimumLength)
		return
	}
	//是否有复杂字符
	if requireComplexity == 1 && !hasComplexCharacters(password) {
		err = errors.New("密码中必须包含复杂字符")
		return
	}
	//是否有数字
	if requireDigit == 1 && !hasDigit(password) {
		err = errors.New("密码必须包含数字")
		return
	}
	//是否有小写字母
	if requireLowercase == 1 && !hasLowercaseLetter(password) {
		err = errors.New("密码必须包含小写字母")
		return
	}
	//是否有大写字母
	if requireUppercase == 1 && !hasUppercaseLetter(password) {
		err = errors.New("密码必须包含大写字母")
		return
	}
	return
}

// hasComplexCharacters：检查字符串中是否有复杂字符（特殊字符）
func hasComplexCharacters(str string) bool {
	specialCharacters := "!@#$%^&*()_+-=[]{}|;:,.<>?~"

	for _, char := range str {
		if contains(specialCharacters, string(char)) {
			return true
		}
	}

	return false
}

// hasDigit：检查字符串中是否有数字
func hasDigit(str string) bool {
	for _, char := range str {
		if isDigit(string(char)) {
			return true
		}
	}

	return false
}

// hasLowercaseLetter：检查字符串中是否有小写字母
func hasLowercaseLetter(str string) bool {
	for _, char := range str {
		if isLowercase(string(char)) {
			return true
		}
	}

	return false
}

// hasUppercaseLetter：检查字符串中是否有大写字母
func hasUppercaseLetter(str string) bool {
	for _, char := range str {
		if isUppercase(string(char)) {
			return true
		}
	}

	return false
}

// isDigit：判断字符是否是数字
func isDigit(c string) bool {
	return c >= "0" && c <= "9"
}

// isLowercase：判断字符是否是小写字母
func isLowercase(c string) bool {
	return c >= "a" && c <= "z"
}

// isUppercase：判断字符是否是大写字母
func isUppercase(c string) bool {
	return c >= "A" && c <= "Z"
}

// contains：判断字符串是否包含指定字符
func contains(str, char string) bool {
	for _, c := range str {
		if string(c) == char {
			return true
		}
	}

	return false
}

// GetMin 返回两个整数中的较小值
func GetMin(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// IsJSONString 判断是否为JSON
func IsJSONString(content string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(content), &js) == nil
}

func InArrayInt(ids []int, id int) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

// 正则表达式用于匹配Base64编码的字符串
var base64Regex = regexp.MustCompile(`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`)

func IsBase64(str string) bool {
	return base64Regex.MatchString(str)
}
