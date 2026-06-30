package upload

import (
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"

	"go-admin/internal/database"
	"go-admin/internal/module/system/model"

	"github.com/gin-gonic/gin"
)

type uploader interface {
	Upload(file *multipart.FileHeader) (string, error)
	Delete(path string) error
	GetURL(path string) string
}

var up uploader

// allowedExts 上传文件扩展名白名单
var allowedExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true, ".zip": true, ".rar": true,
}

// dangerousExts 危险文件扩展名（双重校验）
var dangerousExts = map[string]bool{
	".php": true, ".php3": true, ".php5": true, ".phtml": true,
	".asp": true, ".aspx": true, ".jsp": true, ".jspx": true,
	".sh": true, ".bash": true, ".bat": true, ".cmd": true,
	".exe": true, ".dll": true, ".so": true, ".com": true,
	".js": true, ".vbs": true, ".wsf": true, ".scr": true,
}

func Init() {
	cfg := loadOSSConfig()

	switch cfg.Type {
	case "aliyun":
		ossUploader, err := newAliyunOSS(cfg)
		if err != nil {
			log.Printf("[upload] 阿里云OSS初始化失败，回退到本地存储: %v", err)
			up = &localUploader{}
			return
		}
		up = ossUploader
		log.Printf("[upload] 使用阿里云OSS存储, Bucket: %s", cfg.Bucket)
	case "tencent":
		cosUploader, err := newTencentCOS(cfg)
		if err != nil {
			log.Printf("[upload] 腾讯云COS初始化失败，回退到本地存储: %v", err)
			up = &localUploader{}
			return
		}
		up = cosUploader
		log.Printf("[upload] 使用腾讯云COS存储, Bucket: %s", cfg.Bucket)
	case "minio":
		minioUp, err := newMinIO(cfg)
		if err != nil {
			log.Printf("[upload] MinIO初始化失败，回退到本地存储: %v", err)
			up = &localUploader{}
			return
		}
		up = minioUp
		log.Printf("[upload] 使用MinIO存储, Bucket: %s", cfg.Bucket)
	default:
		up = &localUploader{}
		log.Printf("[upload] 使用本地存储")
	}
}

func loadOSSConfig() OSSConfig {
	var configs []model.SysConfig
	database.DB.Where("config_key LIKE ?", "oss.%").Find(&configs)

	cfgMap := make(map[string]string)
	for _, c := range configs {
		key := c.ConfigKey
		if len(key) > 4 && key[:4] == "oss." {
			cfgMap[key[4:]] = c.Value
		}
	}

	return OSSConfig{
		Type:      cfgMap["type"],
		Endpoint:  cfgMap["endpoint"],
		Bucket:    cfgMap["bucket"],
		AccessKey: cfgMap["access_key"],
		SecretKey: cfgMap["secret_key"],
		Domain:    cfgMap["domain"],
	}
}

func Reload() {
	Init()
}

// ValidateFile 校验文件扩展名白名单
func ValidateFile(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))

	// 双重校验：先检查危险扩展名
	if dangerousExts[ext] {
		return fmt.Errorf("不允许上传 %s 类型文件", ext)
	}

	// 再检查白名单
	if !allowedExts[ext] {
		return fmt.Errorf("不支持的文件类型 %s", ext)
	}

	return nil
}

func Upload(file *multipart.FileHeader) (string, error) {
	if up == nil {
		return "", fmt.Errorf("上传模块未初始化")
	}

	// 校验文件扩展名
	if err := ValidateFile(file.Filename); err != nil {
		return "", err
	}

	return up.Upload(file)
}

func UploadWithContext(c *gin.Context, file *multipart.FileHeader) (string, error) {
	return Upload(file)
}

func Delete(path string) error {
	if up == nil {
		return fmt.Errorf("上传模块未初始化")
	}
	return up.Delete(path)
}

func GetURL(path string) string {
	if up == nil {
		return "/uploads/" + path
	}
	return up.GetURL(path)
}
