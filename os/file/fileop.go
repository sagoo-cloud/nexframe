package file

import (
	"bufio"
	"github.com/sagoo-cloud/nexframe/utils/gstr"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"context"
	"golang.org/x/sync/semaphore"
)

const (
	Separator = string(filepath.Separator)
)

// FileInfoCache 用于缓存文件信息
type FileInfoCache struct {
	info     os.FileInfo
	lastRead time.Time
}

// FileCache 全局文件信息缓存
var FileCache = struct {
	sync.RWMutex
	m map[string]*FileInfoCache
}{m: make(map[string]*FileInfoCache)}

// getCachedFileInfo 获取缓存的文件信息，如果缓存过期则重新读取
func getCachedFileInfo(fullPath string, forceRefresh bool) (os.FileInfo, error) {
	if !forceRefresh {
		FileCache.RLock()
		cache, ok := FileCache.m[fullPath]
		FileCache.RUnlock()

		if ok && time.Since(cache.lastRead) < time.Minute {
			return cache.info, nil
		}
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	FileCache.Lock()
	FileCache.m[fullPath] = &FileInfoCache{info: info, lastRead: time.Now()}
	FileCache.Unlock()

	return info, nil
}

// RefreshFileCache 强制刷新文件缓存
func RefreshFileCache(fullPath string) error {
	_, err := getCachedFileInfo(fullPath, true)
	return err
}

// Filer 文件接口
type Filer interface {
	Size() (int64, error)
	ReadAll() ([]byte, error)
	ReadLine() (string, error)
	ReadBlock(size int64) ([]byte, int64, error)
	Write(b []byte) error
	WriteAt(b []byte, offset int64) (int, error)
	WriteAppend(b []byte) error
	Truncate() error
	Ext() string
	Name() string
	Path() string
	FullPath() string
	ModifyTime() (time.Time, error)
	StreamRead(handler func([]byte) error) error
	ChunkRead(chunkSize int64, handler func([]byte) error) error
}

// File 文件类
type File struct {
	name     string
	path     string
	fullPath string
	rwMu     sync.RWMutex
}

// NewFile 新建文件
func NewFile(fullPath string) Filer {
	return &File{
		name:     filepath.Base(fullPath),
		path:     filepath.Dir(fullPath),
		fullPath: fullPath,
	}
}

// Size 获得文件大小
func (f *File) Size() (int64, error) {
	info, err := getCachedFileInfo(f.fullPath, false)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ReadAll 读取全部
func (f *File) ReadAll() ([]byte, error) {
	f.rwMu.RLock()
	defer f.rwMu.RUnlock()
	return os.ReadFile(f.fullPath)
}

// ReadLine 读取一行
func (f *File) ReadLine() (string, error) {
	f.rwMu.RLock()
	defer f.rwMu.RUnlock()
	file, err := os.Open(f.fullPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", scanner.Err()
}

// ReadBlock 读取块
func (f *File) ReadBlock(size int64) ([]byte, int64, error) {
	f.rwMu.RLock()
	defer f.rwMu.RUnlock()
	file, err := os.Open(f.fullPath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	buf := make([]byte, size)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, 0, err
	}
	return buf[:n], int64(n), nil
}

// StreamRead 流式读取文件
func (f *File) StreamRead(handler func([]byte) error) error {
	f.rwMu.RLock()
	defer f.rwMu.RUnlock()
	file, err := os.Open(f.fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if err := handler(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}

// ChunkRead 分块读取大文件
func (f *File) ChunkRead(chunkSize int64, handler func([]byte) error) error {
	f.rwMu.RLock()
	defer f.rwMu.RUnlock()
	file, err := os.Open(f.fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, chunkSize)
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if err := handler(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}

// Write 写入
func (f *File) Write(b []byte) error {
	f.rwMu.Lock()
	defer f.rwMu.Unlock()
	err := os.WriteFile(f.fullPath, b, 0644)
	if err == nil {
		RefreshFileCache(f.fullPath)
	}
	return err
}

// WriteAt 在固定点写入
func (f *File) WriteAt(b []byte, offset int64) (int, error) {
	f.rwMu.Lock()
	defer f.rwMu.Unlock()
	file, err := os.OpenFile(f.fullPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	n, err := file.WriteAt(b, offset)
	if err == nil {
		RefreshFileCache(f.fullPath)
	}
	return n, err
}

// WriteAppend 追加文件
func (f *File) WriteAppend(b []byte) error {
	f.rwMu.Lock()
	defer f.rwMu.Unlock()
	file, err := os.OpenFile(f.fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(b)
	if err == nil {
		RefreshFileCache(f.fullPath)
	}
	return err
}

// Truncate 清空文件
func (f *File) Truncate() error {
	f.rwMu.Lock()
	defer f.rwMu.Unlock()
	err := os.Truncate(f.fullPath, 0)
	if err == nil {
		RefreshFileCache(f.fullPath)
	}
	return err
}

// Ext 文件后缀名
func (f *File) Ext() string {
	return filepath.Ext(f.fullPath)
}

// Name 文件名
func (f *File) Name() string {
	return f.name
}

// Path 文件路径
func (f *File) Path() string {
	return f.path
}

// FullPath 文件完整路径
func (f *File) FullPath() string {
	return f.fullPath
}

// ModifyTime 文件修改时间
func (f *File) ModifyTime() (time.Time, error) {
	info, err := getCachedFileInfo(f.fullPath, false)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// ProcessFilesInParallel 并行处理多个文件，带并发限制
func ProcessFilesInParallel(files []string, processor func(string) error, maxConcurrency int64) error {
	ctx := context.Background()
	sem := semaphore.NewWeighted(maxConcurrency)
	var wg sync.WaitGroup
	errChan := make(chan error, len(files))

	for _, file := range files {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}

		wg.Add(1)
		go func(f string) {
			defer sem.Release(1)
			defer wg.Done()
			if err := processor(f); err != nil {
				errChan <- err
			}
		}(file)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// ChunkProcessor 用于处理大文件的分块处理器
type ChunkProcessor struct {
	ChunkSize int64
	Handler   func([]byte) error
}

// ProcessLargeFile 处理大文件
func ProcessLargeFile(filePath string, processor ChunkProcessor) error {
	file := NewFile(filePath)
	return file.ChunkRead(processor.ChunkSize, processor.Handler)
}

// RefreshAllFileCache 刷新所有文件缓存
func RefreshAllFileCache() {
	FileCache.Lock()
	defer FileCache.Unlock()
	for path := range FileCache.m {
		delete(FileCache.m, path)
	}
}

// SetCacheExpiration 设置缓存过期时间
func SetCacheExpiration(duration time.Duration) {
	// 实现缓存过期时间的设置逻辑
}

// InitFileSystem 初始化文件系统
func InitFileSystem() {
	// 可以在这里进行一些初始化操作，比如设置缓存过期时间
	SetCacheExpiration(time.Minute * 5)
}

func Join(paths ...string) string {
	var s string
	for _, path := range paths {
		if s != "" {
			s += Separator
		}
		s += gstr.TrimRight(path, Separator)
	}
	return s
}

func Temp(names ...string) string {
	path := os.TempDir()
	for _, name := range names {
		path = Join(path, name)
	}
	return path
}
