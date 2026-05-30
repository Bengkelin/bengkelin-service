package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/config"
	applog "github.com/Bengkelin/bengkelin-service/internal/log"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryProvider struct {
	client *cloudinary.Cloudinary
	config config.StorageCloudinary
}

func NewCloudinaryProvider() (*CloudinaryProvider, error) {
	cfg := config.GetConfig().Cloudinary

	applog.Info("Initializing Cloudinary provider",
		"cloud_name", cfg.CloudName,
		"upload_folder", cfg.UploadFolder,
		"has_url", cfg.URL != "",
	)

	var cld *cloudinary.Cloudinary
	var err error

	if cfg.URL != "" {
		applog.Info("Using CLOUDINARY_URL for authentication")
		cld, err = cloudinary.NewFromURL(cfg.URL)
	} else {
		applog.Info("Using individual params for authentication",
			"api_key", cfg.ApiKey[:4]+"***",
		)
		cld, err = cloudinary.NewFromParams(cfg.CloudName, cfg.ApiKey, cfg.ApiSecret)
	}
	if err != nil {
		applog.Error("Failed to initialize Cloudinary client", "error", err)
		return nil, fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	applog.Info("Cloudinary provider initialized successfully")
	return &CloudinaryProvider{
		client: cld,
		config: cfg,
	}, nil
}

func (p *CloudinaryProvider) Upload(file *multipart.FileHeader, cfg FileUploadConfig) (*UploadResult, error) {
	applog.Info("Cloudinary upload started",
		"filename", file.Filename,
		"size_bytes", file.Size,
		"url_path", cfg.URLPath,
	)

	folder := p.config.UploadFolder
	if folder == "" {
		folder = "bengkelin"
	}

	originalName := strings.TrimSuffix(filepath.Base(file.Filename), filepath.Ext(file.Filename))
	safeName := strings.ReplaceAll(strings.ToLower(originalName), " ", "-")
	publicID := fmt.Sprintf("%s-%d", safeName, time.Now().UnixMilli())

	uploadParams := uploader.UploadParams{
		PublicID:     publicID,
		Folder:       fmt.Sprintf("%s/%s", folder, cfg.URLPath),
		ResourceType: "image",
	}

	applog.Info("Uploading to Cloudinary",
		"public_id", publicID,
		"folder", uploadParams.Folder,
	)

	result, err := p.client.Upload.Upload(context.Background(), file, uploadParams)
	if err != nil {
		applog.Error("Cloudinary upload failed",
			"error", err,
			"filename", file.Filename,
			"public_id", publicID,
		)
		return nil, fmt.Errorf("failed to upload to Cloudinary: %w", err)
	}

	// Check for API-level errors in the response
	if result.Error.Message != "" {
		applog.Error("Cloudinary API returned error",
			"api_error", result.Error.Message,
			"filename", file.Filename,
			"public_id", publicID,
		)
		return nil, fmt.Errorf("cloudinary API error: %s", result.Error.Message)
	}

	applog.Info("Cloudinary upload successful",
		"public_id", result.PublicID,
		"secure_url", result.SecureURL,
		"bytes", result.Bytes,
		"format", result.Format,
		"width", result.Width,
		"height", result.Height,
	)

	return &UploadResult{
		URL:      result.SecureURL,
		Filename: filepath.Base(result.SecureURL),
		Size:     int64(result.Bytes),
	}, nil
}

func (p *CloudinaryProvider) Delete(url string) error {
	publicID := extractPublicID(url)
	if publicID == "" {
		return fmt.Errorf("could not extract public ID from URL: %s", url)
	}

	applog.Info("Deleting from Cloudinary", "public_id", publicID, "url", url)

	_, err := p.client.Upload.Destroy(context.Background(), uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		applog.Error("Failed to delete from Cloudinary", "error", err, "public_id", publicID)
		return fmt.Errorf("failed to delete from Cloudinary: %w", err)
	}

	applog.Info("Deleted from Cloudinary", "public_id", publicID)
	return nil
}

// extractPublicID extracts the Cloudinary public ID from a secure URL.
// URL format: https://res.cloudinary.com/{cloud}/image/upload/v{ver}/{public_id}.{ext}
func extractPublicID(url string) string {
	parts := strings.Split(url, "/upload/")
	if len(parts) < 2 {
		return ""
	}
	// Remove version prefix if present (e.g., "v1234567890/")
	afterUpload := parts[1]
	if idx := strings.Index(afterUpload, "/"); idx != -1 && strings.HasPrefix(afterUpload, "v") {
		afterUpload = afterUpload[idx+1:]
	}
	// Remove file extension
	publicID := strings.TrimSuffix(afterUpload, filepath.Ext(afterUpload))
	return publicID
}
