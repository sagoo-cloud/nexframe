package meta

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// FileUploadMeta 文件上传元数据结构
type FileUploadMeta struct {
	FileName    string                `json:"fileName"`    // 文件名
	Size        int64                 `json:"size"`        // 文件大小
	ContentType string                `json:"contentType"` // 文件类型
	FileHeader  *multipart.FileHeader `json:"-"`           // 文件头信息（不输出到JSON）
}

// DetectContentType 检测文件的实际MIME类型
func (f *FileUploadMeta) DetectContentType() error {
	if f.FileHeader == nil {
		return fmt.Errorf("文件头信息不存在")
	}

	file, err := f.FileHeader.Open()
	if err != nil {
		return fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	// 读取文件头部用于类型检测
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 检测内容类型
	f.ContentType = http.DetectContentType(buffer[:n])

	// 重置文件指针到开始位置
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("重置文件指针失败: %w", err)
	}

	return nil
}

// GetFile 获取上传的文件内容
func (f *FileUploadMeta) GetFile() (multipart.File, error) {
	if f.FileHeader == nil {
		return nil, fmt.Errorf("文件头信息不存在")
	}
	return f.FileHeader.Open()
}

// CopyTo 复制文件到目标 writer
func (f *FileUploadMeta) CopyTo(writer io.Writer) error {
	file, err := f.GetFile()
	if err != nil {
		return fmt.Errorf("获取文件失败: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	return nil
}
