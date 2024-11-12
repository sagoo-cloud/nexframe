// meta/file_upload_meta.go

package meta

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// FileUploadMeta 文件上传元数据结构
type FileUploadMeta struct {
	// 文件原始名称
	FileName string `json:"fileName"`
	// 文件大小（字节）
	Size int64 `json:"size"`
	// 文件MIME类型
	ContentType string `json:"contentType"`
	// 文件头信息（框架内部使用，不输出到JSON）
	FileHeader *multipart.FileHeader `json:"-"`
}

// GetFile 获取上传的文件对象
func (f *FileUploadMeta) GetFile() (multipart.File, error) {
	if f.FileHeader == nil {
		return nil, fmt.Errorf("未上传文件")
	}
	return f.FileHeader.Open()
}

// SaveTo 将上传的文件保存到指定路径
// path: 保存文件的完整路径（包括文件名）
func (f *FileUploadMeta) SaveTo(path string) error {
	if f.FileHeader == nil {
		return fmt.Errorf("未上传文件")
	}

	// 打开源文件
	src, err := f.FileHeader.Open()
	if err != nil {
		return fmt.Errorf("无法打开上传的文件: %w", err)
	}
	defer src.Close()

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建目标文件
	dst, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, src); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}

	return nil
}

// DetectContentType 检测文件的实际MIME类型
func (f *FileUploadMeta) DetectContentType() error {
	if f.FileHeader == nil {
		return fmt.Errorf("未上传文件")
	}

	file, err := f.FileHeader.Open()
	if err != nil {
		return fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	// 读取文件头部用于类型检测
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取文件头部失败: %w", err)
	}

	// 检测内容类型
	f.ContentType = http.DetectContentType(buffer)
	return nil
}
