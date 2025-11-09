package helper

import (
	"fmt"
	"mime/multipart"
	"strings"
)

const (
	MaxImageSize = 5 * 1024 * 1024 // 5MB
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

type FileHelper struct{}

func NewFileHelper() *FileHelper {
	return &FileHelper{}
}

func (h *FileHelper) ValidateImageFile(file *multipart.FileHeader) error {
	if file == nil {
		return fmt.Errorf("ファイルが提供されていません")
	}

	if file.Size > MaxImageSize {
		return fmt.Errorf("ファイルサイズが大きすぎます。最大5MBまでです")
	}

	contentType := file.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return fmt.Errorf("サポートされていないファイル形式です。JPEG、PNG、GIF、WebPのみサポートされています")
	}

	return nil
}

func (h *FileHelper) GetFileExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		parts := strings.Split(contentType, "/")
		if len(parts) == 2 {
			return "." + parts[1]
		}
		return ""
	}
}
