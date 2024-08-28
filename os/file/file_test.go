package file

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileHandler(t *testing.T) {
	tempDir := t.TempDir()
	config := UploadConfig{
		Dir:        tempDir,
		Format:     "2006/01/02/",
		MaxSize:    1024 * 1024,
		AllowedExt: []string{".jpg", ".png", ".gif"},
	}

	handler, err := NewFileHandler(config)
	if err != nil {
		t.Fatalf("Failed to create FileHandler: %v", err)
	}

	if handler.cfg.Dir != tempDir {
		t.Errorf("Expected Dir %s, got %s", tempDir, handler.cfg.Dir)
	}
}

func TestUpload(t *testing.T) {
	tempDir := t.TempDir()
	config := UploadConfig{
		Dir:        tempDir,
		Format:     "2006/01/02/",
		MaxSize:    1024 * 1024,
		AllowedExt: []string{".jpg", ".png", ".gif"},
	}

	handler, _ := NewFileHandler(config)

	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		results, err := handler.Upload("file", r)
		if err != nil {
			t.Fatalf("Upload failed: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		if results[0].Error != nil {
			t.Fatalf("Upload result contains error: %v", results[0].Error)
		}
		fmt.Fprint(w, results[0].Filename)
	}))
	defer ts.Close()

	// Create a test file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.jpg")
	img := createTestImage()
	jpeg.Encode(part, img, nil)
	writer.Close()

	// Send request
	resp, err := http.Post(ts.URL, writer.FormDataContentType(), body)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	uploadedFilename, _ := io.ReadAll(resp.Body)
	if _, err := os.Stat(filepath.Join(tempDir, string(uploadedFilename))); os.IsNotExist(err) {
		t.Fatalf("Uploaded file does not exist")
	}
}

func TestWatermark(t *testing.T) {
	tempDir := t.TempDir()
	watermarkPath := filepath.Join(tempDir, "watermark.png")
	createTestWatermark(watermarkPath)

	config := &WatermarkConfig{
		Path:    watermarkPath,
		Padding: 10,
		Pos:     BottomRight,
	}

	watermark, err := NewWatermark(config)
	if err != nil {
		t.Fatalf("Failed to create Watermark: %v", err)
	}

	testCases := []struct {
		name     string
		ext      string
		createFn func() image.Image
		encodeFn func(io.Writer, image.Image) error
	}{
		{"JPEG", ".jpg", createTestImage, func(w io.Writer, img image.Image) error { return jpeg.Encode(w, img, nil) }},
		{"PNG", ".png", createTestImage, png.Encode},
		{"GIF", ".gif", createTestGIF, func(w io.Writer, img image.Image) error {
			return gif.Encode(w, img, nil)
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test image file
			imgPath := filepath.Join(tempDir, "test"+tc.ext)
			imgFile, _ := os.Create(imgPath)
			img := tc.createFn()
			tc.encodeFn(imgFile, img)
			imgFile.Close()

			// Apply watermark
			imgFile, _ = os.OpenFile(imgPath, os.O_RDWR, 0644)
			err := watermark.Mark(imgFile, tc.ext)
			imgFile.Close()
			if err != nil {
				t.Fatalf("Failed to apply watermark: %v", err)
			}

			// Verify watermark was applied
			imgFile, _ = os.Open(imgPath)
			defer imgFile.Close()
			watermarkedImg, _, err := image.Decode(imgFile)
			if err != nil {
				t.Fatalf("Failed to decode watermarked image: %v", err)
			}

			// Check if the watermark is present in the bottom right corner
			bounds := watermarkedImg.Bounds()
			watermarkBounds := watermark.image.Bounds()
			checkRegion := image.Rect(
				bounds.Max.X-watermarkBounds.Dx()-config.Padding,
				bounds.Max.Y-watermarkBounds.Dy()-config.Padding,
				bounds.Max.X-config.Padding,
				bounds.Max.Y-config.Padding,
			)

			isDifferent := false
			for y := checkRegion.Min.Y; y < checkRegion.Max.Y; y++ {
				for x := checkRegion.Min.X; x < checkRegion.Max.X; x++ {
					if !colorEquals(watermarkedImg.At(x, y), img.At(x, y)) {
						isDifferent = true
						break
					}
				}
				if isDifferent {
					break
				}
			}

			if !isDifferent {
				t.Errorf("Watermark not detected in the expected region")
			}
		})
	}
}

func colorEquals(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
func createTestImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	blue := color.RGBA{0, 0, 255, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)
	return img
}

func createTestGIF() image.Image {
	img := image.NewPaletted(image.Rect(0, 0, 100, 100), color.Palette{color.White, color.Black})
	for i := 0; i < 100; i += 10 {
		for j := 0; j < 100; j += 10 {
			img.Set(i, j, color.Black)
		}
	}
	return img
}

func createTestWatermark(path string) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	red := color.RGBA{255, 0, 0, 128}
	draw.Draw(img, img.Bounds(), &image.Uniform{red}, image.ZP, draw.Src)
	f, _ := os.Create(path)
	defer f.Close()
	png.Encode(f, img)
}

func TestIsAllowedExt(t *testing.T) {
	handler := &FileHandler{
		cfg: UploadConfig{
			AllowedExt: []string{".jpg", ".png", ".gif"},
		},
	}

	testCases := []struct {
		ext      string
		expected bool
	}{
		{".jpg", true},
		{".JPG", true},
		{".png", true},
		{".gif", true},
		{".bmp", false},
		{".txt", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.ext, func(t *testing.T) {
			result := handler.isAllowedExt(tc.ext)
			if result != tc.expected {
				t.Errorf("For extension %s, expected %v but got %v", tc.ext, tc.expected, result)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	handler := &FileHandler{}

	testCases := []struct {
		original string
		ext      string
	}{
		{"test.jpg", ".jpg"},
		{"file with spaces.png", ".png"},
		{"日本語.gif", ".gif"},
	}

	for _, tc := range testCases {
		t.Run(tc.original, func(t *testing.T) {
			result := handler.generateFilename(tc.original)
			if filepath.Ext(result) != tc.ext {
				t.Errorf("Expected extension %s, got %s", tc.ext, filepath.Ext(result))
			}
			if len(result) != 64+len(tc.ext) {
				t.Errorf("Expected filename length %d, got %d", 64+len(tc.ext), len(result))
			}
		})
	}
}
