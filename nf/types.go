package nf

import (
	"mime/multipart"
)

// UploadedFile 表示上传的文件
type UploadedFile struct {
	Filename    string
	Size        int64
	ContentType string
	File        multipart.File
	FileHeader  *multipart.FileHeader
}

// FileField 定义文件字段的配置
type FileField struct {
	MaxSize    int64    // 最大文件大小（字节）
	AllowTypes []string // 允许的文件类型
	Required   bool     // 是否必需
}

// FileConfig 文件上传配置
type FileConfig struct {
	MaxFileSize  int64                // 单个文件最大大小
	MaxTotalSize int64                // 总上传大小限制
	AllowedTypes []string             // 允许的文件类型
	SavePath     string               // 文件保存路径
	Fields       map[string]FileField // 字段配置
}

// DefaultFileConfig 默认文件上传配置
var DefaultFileConfig = FileConfig{
	MaxFileSize:  10 << 20, // 10MB
	MaxTotalSize: 50 << 20, // 50MB
	AllowedTypes: []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	},
	SavePath: "uploads",
}
