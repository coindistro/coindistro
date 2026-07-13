package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// File represents a stored file.
type File struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	MIMEType   string    `json:"mime_type"`
	MD5Hash    string    `json:"md5_hash,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
	PublicURL  string    `json:"public_url,omitempty"`
}

// Provider defines the interface for file storage operations.
type Provider interface {
	// Upload uploads a file to storage.
	Upload(ctx context.Context, path string, content io.Reader, opts ...UploadOption) (*File, error)
	// Download downloads a file from storage.
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	// Delete deletes a file from storage.
	Delete(ctx context.Context, path string) error
	// Exists checks if a file exists in storage.
	Exists(ctx context.Context, path string) (bool, error)
	// List lists files in a directory.
	List(ctx context.Context, prefix string) ([]File, error)
	// URL returns the public URL for a file.
	URL(ctx context.Context, path string) (string, error)
	// Close closes the storage provider connection.
	Close() error
}

// UploadOptions holds options for file uploads.
type UploadOptions struct {
	MIMEType  string
	Overwrite bool
	ACL       string
	Metadata  map[string]string
}

// UploadOption is a function that sets upload options.
type UploadOption func(*UploadOptions)

// WithMIMEType sets the MIME type for the upload.
func WithMIMEType(mimeType string) UploadOption {
	return func(opts *UploadOptions) {
		opts.MIMEType = mimeType
	}
}

// WithOverwrite allows overwriting an existing file.
func WithOverwrite(overwrite bool) UploadOption {
	return func(opts *UploadOptions) {
		opts.Overwrite = overwrite
	}
}

// WithACL sets the access control for the upload.
func WithACL(acl string) UploadOption {
	return func(opts *UploadOptions) {
		opts.ACL = acl
	}
}

// WithMetadata sets metadata for the upload.
func WithMetadata(metadata map[string]string) UploadOption {
	return func(opts *UploadOptions) {
		opts.Metadata = metadata
	}
}

// LocalProvider stores files on the local filesystem.
type LocalProvider struct {
	basePath string
	baseURL  string
	logger   *zap.Logger
}

// NewLocalProvider creates a new local storage provider.
func NewLocalProvider(basePath string, baseURL string, logger *zap.Logger) (*LocalProvider, error) {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("invalid base path: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalProvider{
		basePath: absPath,
		baseURL:  baseURL,
		logger:   logger,
	}, nil
}

// Upload stores a file on the local filesystem.
func (p *LocalProvider) Upload(ctx context.Context, path string, content io.Reader, opts ...UploadOption) (*File, error) {
	options := &UploadOptions{}
	for _, opt := range opts {
		opt(options)
	}

	fullPath := filepath.Join(p.basePath, path)
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Read content to determine size
	data, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	mimeType := options.MIMEType
	if mimeType == "" {
		mimeType = detectMIMEType(path)
	}

	publicURL := ""
	if p.baseURL != "" {
		publicURL = fmt.Sprintf("%s/%s", p.baseURL, path)
	}

	file := &File{
		Name:       filepath.Base(path),
		Path:       path,
		Size:       fileInfo.Size(),
		MIMEType:   mimeType,
		UploadedAt: time.Now().UTC(),
		PublicURL:  publicURL,
	}

	p.logger.Info("file uploaded",
		zap.String("path", path),
		zap.Int64("size", file.Size),
	)

	return file, nil
}

// Download retrieves a file from the local filesystem.
func (p *LocalProvider) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(p.basePath, path)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	return os.Open(fullPath)
}

// Delete removes a file from the local filesystem.
func (p *LocalProvider) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(p.basePath, path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	p.logger.Info("file deleted", zap.String("path", path))
	return nil
}

// Exists checks if a file exists on the local filesystem.
func (p *LocalProvider) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(p.basePath, path)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// List lists files in a directory.
func (p *LocalProvider) List(ctx context.Context, prefix string) ([]File, error) {
	dirPath := filepath.Join(p.basePath, prefix)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []File{}, nil
		}
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	files := make([]File, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		relPath := filepath.Join(prefix, entry.Name())
		publicURL := ""
		if p.baseURL != "" {
			publicURL = fmt.Sprintf("%s/%s", p.baseURL, relPath)
		}

		files = append(files, File{
			Name:       entry.Name(),
			Path:       relPath,
			Size:       info.Size(),
			MIMEType:   detectMIMEType(entry.Name()),
			UploadedAt: info.ModTime().UTC(),
			PublicURL:  publicURL,
		})
	}

	return files, nil
}

// URL returns the public URL for a file.
func (p *LocalProvider) URL(ctx context.Context, path string) (string, error) {
	if p.baseURL == "" {
		return "", fmt.Errorf("base URL not configured")
	}
	return fmt.Sprintf("%s/%s", p.baseURL, path), nil
}

// Close implements the Provider interface.
func (p *LocalProvider) Close() error {
	return nil
}

func detectMIMEType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".csv":
		return "text/csv"
	case ".json":
		return "application/json"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

// InMemoryProvider stores files in memory (useful for testing).
type InMemoryProvider struct {
	files  map[string][]byte
	logger *zap.Logger
}

// NewInMemoryProvider creates a new in-memory storage provider.
func NewInMemoryProvider(logger *zap.Logger) *InMemoryProvider {
	return &InMemoryProvider{
		files:  make(map[string][]byte),
		logger: logger,
	}
}

// Upload stores a file in memory.
func (p *InMemoryProvider) Upload(ctx context.Context, path string, content io.Reader, opts ...UploadOption) (*File, error) {
	data, err := io.ReadAll(content)
	if err != nil {
		return nil, err
	}

	p.files[path] = data

	return &File{
		Name:       filepath.Base(path),
		Path:       path,
		Size:       int64(len(data)),
		UploadedAt: time.Now().UTC(),
	}, nil
}

// Download retrieves a file from memory.
func (p *InMemoryProvider) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	data, ok := p.files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

// Delete removes a file from memory.
func (p *InMemoryProvider) Delete(ctx context.Context, path string) error {
	delete(p.files, path)
	return nil
}

// Exists checks if a file exists in memory.
func (p *InMemoryProvider) Exists(ctx context.Context, path string) (bool, error) {
	_, ok := p.files[path]
	return ok, nil
}

// List lists files with a given prefix.
func (p *InMemoryProvider) List(ctx context.Context, prefix string) ([]File, error) {
	var files []File
	for path := range p.files {
		if len(prefix) == 0 || len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			files = append(files, File{
				Name: filepath.Base(path),
				Path: path,
			})
		}
	}
	return files, nil
}

// URL returns an empty string for in-memory provider.
func (p *InMemoryProvider) URL(ctx context.Context, path string) (string, error) {
	return "", nil
}

// Close implements the Provider interface.
func (p *InMemoryProvider) Close() error {
	p.files = nil
	return nil
}
