package nf

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

// NewStaticHandler 创建静态文件处理器
func (f *APIFramework) NewStaticHandler(fs fs.FS, dir string) *StaticHandler {
	return &StaticHandler{FS: fs, Directory: dir}
}

// SetFileSystem 设置文件系统
func (f *APIFramework) SetFileSystem(fs fs.FS) *APIFramework {
	f.fileSystem = fs
	return f
}

// StaticHandler 处理静态文件的结构
type StaticHandler struct {
	FS        fs.FS
	Directory string
}

// ServeHTTP 实现 http.Handler 接口
func (sh *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(sh.Directory, strings.TrimPrefix(r.URL.Path, "/"+sh.Directory+"/"))
	file, err := sh.FS.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if stat.IsDir() {
		http.NotFound(w, r)
		return
	}

	// 设置 Content-Type
	contentType := getContentType(path)
	w.Header().Set("Content-Type", contentType)

	// 设置 Content-Length
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	// 如果是 GET 请求，发送文件内容
	if r.Method == "GET" {
		_, err = io.Copy(w, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
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
