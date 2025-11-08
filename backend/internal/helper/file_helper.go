package helper

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxFileSize = 5 * 1024 * 1024 // 5MB
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

type FileHelper struct {
	uploadDir string
}

func NewFileHelper(uploadDir string) *FileHelper {
	return &FileHelper{
		uploadDir: uploadDir,
	}
}

// ValidateImageFile validates the uploaded image file
func (f *FileHelper) ValidateImageFile(fileHeader *multipart.FileHeader) error {
	// Check file size
	if fileHeader.Size > MaxFileSize {
		return fmt.Errorf("ファイルサイズが大きすぎます。最大5MBまでです")
	}

	// Check content type
	contentType := fileHeader.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return fmt.Errorf("サポートされていない画像形式です。JPEG, PNG, GIF, WebPのみ対応しています")
	}

	return nil
}

// SaveImageFile saves the uploaded image file and returns the relative path
func (f *FileHelper) SaveImageFile(fileHeader *multipart.FileHeader, subDir string) (string, error) {
	// Create upload directory if it doesn't exist
	uploadPath := filepath.Join(f.uploadDir, subDir)
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return "", fmt.Errorf("アップロードディレクトリの作成に失敗しました: %w", err)
	}

	// Generate unique filename
	filename, err := f.generateUniqueFilename(fileHeader.Filename)
	if err != nil {
		return "", fmt.Errorf("ファイル名の生成に失敗しました: %w", err)
	}

	// Open source file
	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("ファイルのオープンに失敗しました: %w", err)
	}
	defer src.Close()

	// Create destination file
	fullPath := filepath.Join(uploadPath, filename)
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("ファイルの作成に失敗しました: %w", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("ファイルのコピーに失敗しました: %w", err)
	}

	// Return relative path from upload directory
	relativePath := filepath.Join(subDir, filename)
	return relativePath, nil
}

// DeleteImageFile deletes the image file at the given relative path
func (f *FileHelper) DeleteImageFile(relativePath string) error {
	fullPath := filepath.Join(f.uploadDir, relativePath)
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, consider it deleted
		}
		return fmt.Errorf("ファイルの削除に失敗しました: %w", err)
	}
	return nil
}

// generateUniqueFilename generates a unique filename while preserving the extension
func (f *FileHelper) generateUniqueFilename(originalFilename string) (string, error) {
	ext := filepath.Ext(originalFilename)

	// Generate random bytes
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Create filename: randomhex + extension
	randomHex := hex.EncodeToString(randomBytes)
	filename := randomHex + strings.ToLower(ext)

	return filename, nil
}
