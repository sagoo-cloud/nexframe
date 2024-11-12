// Package file 提供文件处理、上传和水印功能
package file

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/sagoo-cloud/nexframe/utils/errors/gerror"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FileError 自定义错误类型
type FileError struct {
	Op  string
	Err error
}

func (e *FileError) Error() string {
	return fmt.Sprintf("file %s: %v", e.Op, e.Err)
}

// WatermarkPos 表示水印的位置
type WatermarkPos int

// 水印的位置常量
const (
	TopLeft WatermarkPos = iota
	TopRight
	BottomLeft
	BottomRight
	Center
)

// 允许的图片扩展名
var allowedImageExts = []string{
	".gif", ".jpg", ".jpeg", ".png",
}

// WatermarkConfig 水印配置
type WatermarkConfig struct {
	Path         string
	Padding      int
	Pos          WatermarkPos
	Transparency uint8
	Rotation     float64
}

// UploadConfig 上传配置
type UploadConfig struct {
	Dir        string
	Format     string
	MaxSize    int64
	AllowedExt []string
	Watermark  *WatermarkConfig
}

// FileHandler 处理文件上传和水印
type FileHandler struct {
	fs        fs.FS
	cfg       UploadConfig
	watermark *Watermark
}

// Watermark 用于给图片添加水印功能
type Watermark struct {
	image  image.Image
	gifImg *gif.GIF
	config WatermarkConfig
}

// NewFileHandler 创建一个新的 FileHandler 实例
func NewFileHandler(cfg UploadConfig) (*FileHandler, error) {
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, &FileError{"create dir", err}
	}

	fh := &FileHandler{
		fs:  os.DirFS(cfg.Dir),
		cfg: cfg,
	}

	if cfg.Watermark != nil {
		wm, err := NewWatermark(cfg.Watermark)
		if err != nil {
			return nil, err
		}
		fh.watermark = wm
	}

	return fh, nil
}

// NewWatermark 创建一个新的 Watermark 实例
func NewWatermark(cfg *WatermarkConfig) (*Watermark, error) {
	f, err := os.Open(cfg.Path)
	if err != nil {
		return nil, &FileError{"open watermark", err}
	}
	defer f.Close()

	var img image.Image
	var gifImg *gif.GIF

	switch strings.ToLower(filepath.Ext(cfg.Path)) {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(f)
	case ".png":
		img, err = png.Decode(f)
	case ".gif":
		gifImg, err = gif.DecodeAll(f)
		if err == nil {
			img = gifImg.Image[0]
		}
	default:
		return nil, &FileError{"decode", fmt.Errorf("unsupported image type: %s", filepath.Ext(cfg.Path))}
	}

	if err != nil {
		return nil, &FileError{"decode", err}
	}

	return &Watermark{
		image:  img,
		gifImg: gifImg,
		config: *cfg,
	}, nil
}

// UploadResult 表示单个文件的上传结果
type UploadResult struct {
	Filename string
	Error    error
}

// Upload 执行上传操作
func (fh *FileHandler) Upload(field string, r *http.Request) ([]UploadResult, error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, &FileError{"parse form", err}
	}

	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil, &FileError{"no file", fmt.Errorf("no file uploaded")}
	}

	heads := r.MultipartForm.File[field]
	if len(heads) == 0 {
		return nil, &FileError{"no file", fmt.Errorf("no file in field %s", field)}
	}

	results := make([]UploadResult, len(heads))
	var wg sync.WaitGroup

	for i, head := range heads {
		wg.Add(1)
		go func(i int, head *multipart.FileHeader) {
			defer wg.Done()
			filePath, err := fh.moveFile(head)
			results[i] = UploadResult{Filename: filePath, Error: err}
		}(i, head)
	}

	wg.Wait()
	return results, nil
}

func (fh *FileHandler) moveFile(head *multipart.FileHeader) (string, error) {
	if head.Size > fh.cfg.MaxSize {
		return "", &FileError{"size check", fmt.Errorf("file size exceeds limit")}
	}

	ext := strings.ToLower(filepath.Ext(head.Filename))
	if !fh.isAllowedExt(ext) {
		return "", &FileError{"ext check", fmt.Errorf("file extension not allowed")}
	}

	srcFile, err := head.Open()
	if err != nil {
		return "", &FileError{"open source", err}
	}
	defer srcFile.Close()

	relDir := time.Now().Format(fh.cfg.Format)
	dir := filepath.Join(fh.cfg.Dir, relDir)
	if err = os.MkdirAll(dir, 0755); err != nil {
		return "", &FileError{"create dir", err}
	}

	filename := fh.generateFilename(head.Filename)
	destPath := filepath.Join(dir, filename)

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", &FileError{"create dest", err}
	}
	defer destFile.Close()

	if _, err = io.Copy(destFile, srcFile); err != nil {
		return "", &FileError{"copy", err}
	}

	if fh.watermark != nil && IsAllowedImageExt(ext) {
		if err = fh.watermark.Mark(destFile, ext); err != nil {
			return "", &FileError{"watermark", err}
		}
	}

	return path.Join(relDir, filename), nil
}

func (fh *FileHandler) isAllowedExt(ext string) bool {
	for _, e := range fh.cfg.AllowedExt {
		if strings.EqualFold(e, ext) {
			return true
		}
	}
	return false
}

func (fh *FileHandler) generateFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)
	hash := sha256.Sum256([]byte(name + time.Now().String()))
	return hex.EncodeToString(hash[:]) + ext
}

// Mark 将水印写入文件
// Mark 将水印写入文件
func (w *Watermark) Mark(file *os.File, ext string) error {
	_, err := file.Seek(0, 0)
	if err != nil {
		return &FileError{"seek", err}
	}

	switch strings.ToLower(ext) {
	case ".gif":
		return w.markGIF(file)
	case ".jpg", ".jpeg":
		return w.markJPEG(file)
	case ".png":
		return w.markPNG(file)
	default:
		return &FileError{"mark", fmt.Errorf("unsupported image type: %s", ext)}
	}
}

func (w *Watermark) markJPEG(file *os.File) error {
	// 解码 JPEG 图像
	img, err := jpeg.Decode(file)
	if err != nil {
		return &FileError{"decode jpeg", err}
	}

	// 创建一个新的 RGBA 图像
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// 添加水印
	w.drawWatermark(rgba)

	// 将文件指针移到开头
	if _, err := file.Seek(0, 0); err != nil {
		return &FileError{"seek", err}
	}

	// 将处理后的图像编码为 JPEG 并写入文件
	if err := jpeg.Encode(file, rgba, nil); err != nil {
		return &FileError{"encode jpeg", err}
	}

	return nil
}

func (w *Watermark) markPNG(file *os.File) error {
	// 解码 PNG 图像
	img, err := png.Decode(file)
	if err != nil {
		return &FileError{"decode png", err}
	}

	// 创建一个新的 RGBA 图像
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// 添加水印
	w.drawWatermark(rgba)

	// 将文件指针移到开头
	if _, err := file.Seek(0, 0); err != nil {
		return &FileError{"seek", err}
	}

	// 将处理后的图像编码为 PNG 并写入文件
	if err := png.Encode(file, rgba); err != nil {
		return &FileError{"encode png", err}
	}

	return nil
}

func (w *Watermark) markGIF(file *os.File) error {
	// 解码 GIF 图像
	gifImg, err := gif.DecodeAll(file)
	if err != nil {
		return &FileError{"decode gif", err}
	}

	// 处理每一帧
	for i := range gifImg.Image {
		rgba := image.NewRGBA(gifImg.Image[i].Bounds())
		draw.Draw(rgba, rgba.Bounds(), gifImg.Image[i], gifImg.Image[i].Bounds().Min, draw.Src)
		w.drawWatermark(rgba)
		gifImg.Image[i] = imageToAllocedPaletted(rgba, gifImg.Image[i].Palette)
	}

	// 将文件指针移到开头
	if _, err := file.Seek(0, 0); err != nil {
		return &FileError{"seek", err}
	}

	// 将处理后的 GIF 编码并写入文件
	if err := gif.EncodeAll(file, gifImg); err != nil {
		return &FileError{"encode gif", err}
	}

	return nil
}

// drawWatermark 在给定的 RGBA 图像上绘制水印
func (w *Watermark) drawWatermark(rgba *image.RGBA) {
	bounds := rgba.Bounds()
	watermarkBounds := w.image.Bounds()
	x, y := w.getWatermarkPosition(bounds, watermarkBounds)

	draw.Draw(rgba, image.Rect(x, y, x+watermarkBounds.Dx(), y+watermarkBounds.Dy()),
		w.image, watermarkBounds.Min, draw.Over)
}

// getWatermarkPosition 计算水印的位置
func (w *Watermark) getWatermarkPosition(imgBounds, watermarkBounds image.Rectangle) (x, y int) {
	switch w.config.Pos {
	case TopLeft:
		return w.config.Padding, w.config.Padding
	case TopRight:
		return imgBounds.Dx() - watermarkBounds.Dx() - w.config.Padding, w.config.Padding
	case BottomLeft:
		return w.config.Padding, imgBounds.Dy() - watermarkBounds.Dy() - w.config.Padding
	case BottomRight:
		return imgBounds.Dx() - watermarkBounds.Dx() - w.config.Padding,
			imgBounds.Dy() - watermarkBounds.Dy() - w.config.Padding
	case Center:
		return (imgBounds.Dx() - watermarkBounds.Dx()) / 2,
			(imgBounds.Dy() - watermarkBounds.Dy()) / 2
	default:
		return 0, 0
	}
}

// imageToAllocedPaletted 将 RGBA 图像转换为具有给定调色板的 Paletted 图像
func imageToAllocedPaletted(img *image.RGBA, p color.Palette) *image.Paletted {
	pm := image.NewPaletted(img.Bounds(), p)
	draw.FloydSteinberg.Draw(pm, img.Bounds(), img, image.Point{})
	return pm
}

// IsAllowedImageExt 检查是否是允许的图片扩展名
func IsAllowedImageExt(ext string) bool {
	for _, e := range allowedImageExts {
		if strings.EqualFold(e, ext) {
			return true
		}
	}
	return false
}

// Exists checks whether given `path` exist.
func Exists(path string) bool {
	if stat, err := os.Stat(path); stat != nil && !os.IsNotExist(err) {
		return true
	}
	return false
}

// IsDir checks whether given `path` a directory.
// Note that it returns false if the `path` does not exist.
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// Pwd returns absolute path of current working directory.
// Note that it returns an empty string if retrieving current
// working directory failed.
func Pwd() string {
	path, err := os.Getwd()
	if err != nil {
		return ""
	}
	return path
}

// Chdir changes the current working directory to the named directory.
// If there is an error, it will be of type *PathError.
func Chdir(dir string) (err error) {
	err = os.Chdir(dir)
	if err != nil {
		err = gerror.Wrapf(err, `os.Chdir failed with dir "%s"`, dir)
	}
	return
}

// IsFile checks whether given `path` a file, which means it's not a directory.
// Note that it returns false if the `path` does not exist.
func IsFile(path string) bool {
	s, err := Stat(path)
	if err != nil {
		return false
	}
	return !s.IsDir()
}

// Stat returns a FileInfo describing the named file.
// If there is an error, it will be of type *PathError.
func Stat(path string) (os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		err = gerror.Wrapf(err, `os.Stat failed for file "%s"`, path)
	}
	return info, err
}

// Move renames (moves) `src` to `dst` path.
// If `dst` already exists and is not a directory, it'll be replaced.
func Move(src string, dst string) (err error) {
	err = os.Rename(src, dst)
	if err != nil {
		err = gerror.Wrapf(err, `os.Rename failed from "%s" to "%s"`, src, dst)
	}
	return
}

// Rename is alias of Move.
// See Move.
func Rename(src string, dst string) error {
	return Move(src, dst)
}

// Mkdir creates directories recursively with given `path`.
// The parameter `path` is suggested to be an absolute path instead of relative one.
func Mkdir(path string) (err error) {
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		err = gerror.Wrapf(err, `os.MkdirAll failed for path "%s" with perm "%d"`, path, os.ModePerm)
		return err
	}
	return nil
}
func Basename(path string) string {
	return filepath.Base(path)
}

// Create creates a file with given `path` recursively.
// The parameter `path` is suggested to be absolute path.
func Create(path string) (*os.File, error) {
	dir := Dir(path)
	if !Exists(dir) {
		if err := Mkdir(dir); err != nil {
			return nil, err
		}
	}
	file, err := os.Create(path)
	if err != nil {
		err = gerror.Wrapf(err, `os.Create failed for name "%s"`, path)
	}
	return file, err
}
func Dir(path string) string {
	if path == "." {
		return filepath.Dir(RealPath(path))
	}
	return filepath.Dir(path)
}

// RealPath converts the given `path` to its absolute path
// and checks if the file path exists.
// If the file does not exist, return an empty string.
func RealPath(path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	if !Exists(p) {
		return ""
	}
	return p
}

// Example:
// Ext("main.go")  => .go
// Ext("api.json") => .json
func Ext(path string) string {
	ext := filepath.Ext(path)
	if p := strings.IndexByte(ext, '?'); p != -1 {
		ext = ext[0:p]
	}
	return ext
}
