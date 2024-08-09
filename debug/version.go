package debug

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
)

var (
	binaryVersion    string
	binaryVersionMd5 string
	selfPath         string
	versionOnce      sync.Once
	initOnce         sync.Once
)

func init() {
	initOnce.Do(func() {
		selfPath = getSelfPath()
	})
}

// getSelfPath returns the absolute path of the current running binary.
func getSelfPath() string {
	path, err := exec.LookPath(os.Args[0])
	if err == nil {
		path, err = filepath.Abs(path)
	}
	if err != nil {
		path, _ = filepath.Abs(os.Args[0])
	}
	return path
}

// BinVersion returns the version of the current running binary.
func BinVersion() (string, error) {
	var err error
	versionOnce.Do(func() {
		var binaryContent []byte
		binaryContent, err = os.ReadFile(selfPath)
		if err != nil {
			return
		}
		binaryVersion = strconv.FormatInt(int64(BKDR(binaryContent)), 36)
	})
	return binaryVersion, err
}

// BKDR calculates a hash value using the BKDR hash algorithm.
func BKDR(str []byte) uint32 {
	var (
		seed uint32 = 131 // 31 131 1313 13131 131313 etc..
		hash uint32 = 0
	)
	for i := 0; i < len(str); i++ {
		hash = hash*seed + uint32(str[i])
	}
	return hash
}

// BinVersionMd5 returns the MD5 hash of the current running binary.
func BinVersionMd5() (string, error) {
	var err error
	versionOnce.Do(func() {
		binaryVersionMd5, err = md5File(selfPath)
	})
	return binaryVersionMd5, err
}

// md5File calculates the MD5 hash of a file.
func md5File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
