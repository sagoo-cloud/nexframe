package homedir

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	// ErrCannotExpandUser 表示无法展开用户特定的主目录
	ErrCannotExpandUser = errors.New("cannot expand user-specific home dir")
	// ErrBlankOutput 表示读取主目录时得到空白输出
	ErrBlankOutput = errors.New("blank output when reading home directory")
	// ErrMissingWindowsEnv 表示 Windows 环境变量缺失
	ErrMissingWindowsEnv = errors.New("HOMEDRIVE, HOMEPATH, or USERPROFILE are blank")
)

// disableCache 将禁用主目录的缓存。默认情况下启用缓存。
var disableCache bool

var (
	homedirCache string
	cacheLock    sync.RWMutex
)

// Dir 返回执行用户的主目录。
//
// 它使用特定于操作系统的方法来发现主目录。
// 如果无法检测到主目录，则返回错误。
func Dir() (string, error) {
	if !disableCache {
		cacheLock.RLock()
		cached := homedirCache
		cacheLock.RUnlock()
		if cached != "" {
			return cached, nil
		}
	}

	cacheLock.Lock()
	defer cacheLock.Unlock()

	var result string
	var err error
	if runtime.GOOS == "windows" {
		result, err = dirWindows()
	} else {
		// Unix-like 系统，所以假设为 Unix
		result, err = dirUnix()
	}

	if err != nil {
		return "", err
	}
	homedirCache = result
	return result, nil
}

// Expand 如果路径以 `~` 为前缀，则将路径扩展为包含主目录。
// 如果不以 `~` 为前缀，则按原样返回路径。
func Expand(path string) (string, error) {
	if path == "" || path[0] != '~' {
		return path, nil
	}

	if len(path) > 1 && path[1] != '/' && path[1] != '\\' {
		return "", ErrCannotExpandUser
	}

	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, path[1:]), nil
}

// Reset 清除缓存，强制下一次调用 Dir 重新检测主目录。
// 这通常不需要调用，但在测试中如果通过 HOME 环境变量
// 或其他方式修改主目录时可能会有用。
func Reset() {
	cacheLock.Lock()
	homedirCache = ""
	cacheLock.Unlock()
}

func dirUnix() (string, error) {
	homeEnv := "HOME"
	if runtime.GOOS == "plan9" {
		// 在 plan9 上，环境变量是小写的。
		homeEnv = "home"
	}

	// 首选 HOME 环境变量
	if home := os.Getenv(homeEnv); home != "" {
		return home, nil
	}

	var stdout bytes.Buffer

	// 如果失败，尝试特定于操作系统的命令
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("sh", "-c", `dscl -q . -read /Users/"$(whoami)" NFSHomeDirectory | sed 's/^[^ ]*: //'`)
		cmd.Stdout = &stdout
		if err := cmd.Run(); err == nil {
			if result := strings.TrimSpace(stdout.String()); result != "" {
				return result, nil
			}
		}
	} else {
		cmd := exec.Command("getent", "passwd", strconv.Itoa(os.Getuid()))
		cmd.Stdout = &stdout
		if err := cmd.Run(); err == nil {
			if passwd := strings.TrimSpace(stdout.String()); passwd != "" {
				// username:password:uid:gid:gecos:home:shell
				passwdParts := strings.SplitN(passwd, ":", 7)
				if len(passwdParts) > 5 {
					return passwdParts[5], nil
				}
			}
		}
	}

	// 如果其他方法都失败，尝试 shell
	stdout.Reset()
	cmd := exec.Command("sh", "-c", "cd && pwd")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", ErrBlankOutput
	}

	return result, nil
}

func dirWindows() (string, error) {
	// 首选 HOME 环境变量
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// 优先使用标准环境变量 USERPROFILE
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home, nil
	}

	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		return "", ErrMissingWindowsEnv
	}

	return home, nil
}
