package nf

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe/os/file"
	"net/http"
	"path/filepath"
	"strings"
)

// noListingFileSystem 包装 http.FileSystem，禁止目录列表
type noListingFileSystem struct {
	fs http.FileSystem
}

// Open 重写 Open 方法以禁止目录列表
func (nfs noListingFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			f.Close()
			return nil, err
		}
	}

	return f, nil
}

// SetServerRoot 设置文档根目录用于静态服务
func (f *APIFramework) SetServerRoot(root string) *APIFramework {
	var realPath string
	if p, err := file.Search(root); err != nil {
		fmt.Printf(`SetServerRoot failed: %+v \n`, err)
		realPath = root
	} else {
		realPath = p
	}

	f.wwwRoot = strings.TrimRight(realPath, file.Separator)
	f.config.FileServerEnabled = true
	return f
}

// NewStaticHandler 创建静态文件处理器
func (f *APIFramework) NewStaticHandler(fs http.FileSystem, dir string) http.Handler {
	return http.FileServer(noListingFileSystem{fs})
}

// SetFileSystem 设置文件系统
func (f *APIFramework) SetFileSystem(fs http.FileSystem) *APIFramework {
	f.fileSystem = fs
	return f
}

// getContentType 根据文件扩展名获取 MIME 类型
func getContentType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
