package upload

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"go-admin/config"

	"github.com/google/uuid"
)

type localUploader struct{}

func (l *localUploader) Upload(file *multipart.FileHeader) (string, error) {
	savePath := config.Cfg.Upload.SavePath

	dir := filepath.Join(savePath, time.Now().Format("2006/01/02"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	ext := filepath.Ext(file.Filename)
	storageName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	fullPath := filepath.Join(dir, storageName)

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(src); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	relPath, _ := filepath.Rel(savePath, fullPath)
	relPath = filepath.ToSlash(relPath)
	return relPath, nil
}

func (l *localUploader) Delete(path string) error {
	fullPath := filepath.Join(config.Cfg.Upload.SavePath, path)
	return os.Remove(fullPath)
}

func (l *localUploader) GetURL(path string) string {
	return "/uploads/" + path
}
