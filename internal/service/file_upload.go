package service

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/config"
	applog "github.com/Bengkelin/bengkelin-service/internal/log"
)

// FileUploadConfig defines allowed file types and size limits
type FileUploadConfig struct {
	AllowedExtensions []string
	MaxFileSize       int64 // bytes
	BaseDir           string
	URLPath           string
}

var (
	AvatarUploadConfig = FileUploadConfig{
		AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".webp"},
		MaxFileSize:       5 * 1024 * 1024, // 5MB
		BaseDir:           "public/avatars",
		URLPath:           "avatar",
	}

	PhotoUploadConfig = FileUploadConfig{
		AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".webp"},
		MaxFileSize:       10 * 1024 * 1024, // 10MB
		BaseDir:           "public/bengkels",
		URLPath:           "bengkel",
	}

	VehicleUploadConfig = FileUploadConfig{
		AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".webp"},
		MaxFileSize:       5 * 1024 * 1024, // 5MB
		BaseDir:           "public/vehicles",
		URLPath:           "vehicle",
	}
)

// StorageProvider abstracts file storage backend
type StorageProvider interface {
	Upload(file *multipart.FileHeader, cfg FileUploadConfig) (*UploadResult, error)
	Delete(url string) error
}

// FileUploadService handles file upload operations
type FileUploadService struct {
	provider StorageProvider
}

func NewFileUploadService() *FileUploadService {
	return &FileUploadService{}
}

func NewFileUploadServiceWithProvider(provider StorageProvider) *FileUploadService {
	return &FileUploadService{provider: provider}
}

func NewFileUploadServiceFromConfig() *FileUploadService {
	cfg := config.GetConfig()
	applog.Info("Storage provider configured", "provider", cfg.StorageProvider)

	if cfg.StorageProvider == "cloudinary" {
		provider, err := NewCloudinaryProvider()
		if err != nil {
			applog.Error("Failed to init Cloudinary, falling back to local storage", "error", err)
			return &FileUploadService{}
		}
		applog.Info("Using Cloudinary storage provider")
		return &FileUploadService{provider: provider}
	}

	applog.Info("Using local storage provider")
	return &FileUploadService{}
}

// UploadResult contains the result of a file upload
type UploadResult struct {
	URL      string
	Filename string
	Size     int64
}

// UploadFile uploads a single file and returns the URL
func (s *FileUploadService) UploadFile(file *multipart.FileHeader, cfg FileUploadConfig, protocol, host string) (*UploadResult, error) {
	// Validate file size
	if file.Size > cfg.MaxFileSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed %d bytes", file.Size, cfg.MaxFileSize)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedExtension(ext, cfg.AllowedExtensions) {
		return nil, fmt.Errorf("file extension %s is not allowed, accepted: %v", ext, cfg.AllowedExtensions)
	}

	// Use cloud provider if configured
	if s.provider != nil {
		return s.provider.Upload(file, cfg)
	}

	// Local storage fallback
	return s.uploadLocal(file, cfg, ext, protocol, host)
}

func (s *FileUploadService) uploadLocal(file *multipart.FileHeader, cfg FileUploadConfig, ext, protocol, host string) (*UploadResult, error) {
	// Ensure directory exists
	if err := os.MkdirAll(cfg.BaseDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	originalName := strings.TrimSuffix(filepath.Base(file.Filename), filepath.Ext(file.Filename))
	safeName := strings.ReplaceAll(strings.ToLower(originalName), " ", "-")
	fileName := fmt.Sprintf("%s-%d%s", safeName, time.Now().Unix(), ext)
	savePath := filepath.Join(".", cfg.BaseDir, fileName)

	// Save file
	if err := saveUploadedFile(file, savePath); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Build URL
	serverConfig := config.GetConfig().Server
	reqHost := host
	if serverConfig.DevMode == "false" {
		reqHost = serverConfig.Host
	} else {
		reqHost = serverConfig.Host + ":" + serverConfig.Port
	}

	url := fmt.Sprintf("%s://%s/api/v1/static/%s/%s", protocol, reqHost, cfg.URLPath, fileName)

	return &UploadResult{
		URL:      url,
		Filename: fileName,
		Size:     file.Size,
	}, nil
}

// UploadMultipleFiles uploads multiple files and returns URLs
func (s *FileUploadService) UploadMultipleFiles(files []*multipart.FileHeader, cfg FileUploadConfig, protocol, host string) ([]string, error) {
	var urls []string
	for _, file := range files {
		result, err := s.UploadFile(file, cfg, protocol, host)
		if err != nil {
			return nil, err
		}
		urls = append(urls, result.URL)
	}
	return urls, nil
}

func (s *FileUploadService) DeleteFile(url string) error {
	if s.provider != nil {
		return s.provider.Delete(url)
	}
	return nil
}

func isAllowedExtension(ext string, allowed []string) bool {
	for _, a := range allowed {
		if ext == a {
			return true
		}
	}
	return false
}

func saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	buf := make([]byte, 32*1024)
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if readErr != nil {
			break
		}
	}
	return nil
}
