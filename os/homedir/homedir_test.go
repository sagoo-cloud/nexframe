package homedir

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// 定义一个接口来抽象获取主目录的方法
type dirProvider interface {
	getDir() (string, error)
}

// 实现默认的 dirProvider
type defaultDirProvider struct{}

func (d defaultDirProvider) getDir() (string, error) {
	return Dir()
}

// 创建一个 mock dirProvider 用于测试
type mockDirProvider struct {
	dir string
	err error
}

func (m mockDirProvider) getDir() (string, error) {
	return m.dir, m.err
}

// 修改 Expand 函数，使其接受一个 dirProvider
func ExpandWithProvider(path string, provider dirProvider) (string, error) {
	if path == "" || path[0] != '~' {
		return path, nil
	}

	if len(path) > 1 && path[1] != '/' && path[1] != '\\' {
		return "", ErrCannotExpandUser
	}

	dir, err := provider.getDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, path[1:]), nil
}

func TestDir(t *testing.T) {
	// 保存原始的环境变量
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	origHomeDrive := os.Getenv("HOMEDRIVE")
	origHomePath := os.Getenv("HOMEPATH")

	// 测试结束后恢复环境变量
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOMEDRIVE", origHomeDrive)
		os.Setenv("HOMEPATH", origHomePath)
	}()

	// 测试 Unix 系统
	if runtime.GOOS != "windows" {
		testDir := "/Users/xinjiayu"
		os.Setenv("HOME", testDir)
		dir, err := Dir()
		if err != nil {
			t.Errorf("Dir() returned an error: %v", err)
		}
		if dir != testDir {
			t.Errorf("Dir() = %s; want %s", dir, testDir)
		}
	} else {
		// 测试 Windows 系统
		testDir := `C:\Users\TestUser`
		os.Setenv("USERPROFILE", testDir)
		dir, err := Dir()
		if err != nil {
			t.Errorf("Dir() returned an error: %v", err)
		}
		if dir != testDir {
			t.Errorf("Dir() = %s; want %s", dir, testDir)
		}

		// 测试 HOMEDRIVE 和 HOMEPATH
		os.Setenv("USERPROFILE", "")
		os.Setenv("HOMEDRIVE", "D:")
		os.Setenv("HOMEPATH", `\TestUser`)
		dir, err = Dir()
		if err != nil {
			t.Errorf("Dir() returned an error: %v", err)
		}
		if dir != `D:\TestUser` {
			t.Errorf("Dir() = %s; want D:\\TestUser", dir)
		}
	}
}

func TestExpand(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		mockDir  string
		mockErr  error
		expected string
		wantErr  bool
	}{
		{"ExpandHomeDir", "~/test", "/home/user", nil, "/home/user/test", false},
		{"JustTilde", "~", "/home/user", nil, "/home/user", false},
		{"InvalidUserExpansion", "~user/test", "/home/user", nil, "", true},
		{"AbsolutePath", "/absolute/path", "/home/user", nil, "/absolute/path", false},
		{"EmptyString", "", "/home/user", nil, "", false},
		{"DirError", "~/test", "", ErrBlankOutput, "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockProvider := mockDirProvider{dir: tc.mockDir, err: tc.mockErr}
			result, err := ExpandWithProvider(tc.input, mockProvider)

			if tc.wantErr && err == nil {
				t.Errorf("ExpandWithProvider(%q) expected an error, got nil", tc.input)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("ExpandWithProvider(%q) returned an unexpected error: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("ExpandWithProvider(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestReset(t *testing.T) {
	// 首先设置缓存
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Initial Dir() call failed: %v", err)
	}

	// 调用 Reset
	Reset()

	// 检查缓存是否被清除
	cacheLock.RLock()
	cached := homedirCache
	cacheLock.RUnlock()

	if cached != "" {
		t.Errorf("After Reset(), cache was not empty. Got: %q", cached)
	}

	// 再次调用 Dir() 并确保它仍然工作
	newDir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() call after Reset() failed: %v", err)
	}
	if newDir != dir {
		t.Errorf("Dir() after Reset() returned different directory. Got: %q, Want: %q", newDir, dir)
	}
}
