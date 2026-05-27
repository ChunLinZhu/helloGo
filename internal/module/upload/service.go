// Package upload — 上传业务逻辑
package upload

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"helloGo/internal/config"
	"helloGo/internal/pkg/pagination"
)

// Service 上传业务接口
type Service interface {
	Upload(file *multipart.FileHeader) (*UploadResponse, error)
	UploadChunk(req *ChunkUploadRequest, file *multipart.FileHeader) error
	MergeChunks(req *MergeRequest) (*UploadResponse, error)
	GetByID(id string) (*UploadResponse, error)
	List(page pagination.Pagination) ([]UploadResponse, int64, error)
	Delete(id string) error
	CleanExpired() error
}

// service 上传业务实现
type service struct {
	repo          Repository
	logger        *zap.Logger
	uploadDest    string
	maxSize       int64
	allowedTypes  map[string]bool
	chunkDir      string // 分片临时存储目录
	ttlDays       int    // 文件保留天数
}

// NewService 创建上传业务层
func NewService(repo Repository, logger *zap.Logger, cfg config.UploadConfig) Service {
	// 解析允许的 MIME 类型
	allowed := make(map[string]bool)
	for _, t := range strings.Split(cfg.AllowedTypes, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			allowed[t] = true
		}
	}

	// 确保上传目录存在
	_ = os.MkdirAll(cfg.Dest, 0755)

	// 分片临时目录
	chunkDir := filepath.Join(cfg.Dest, ".chunks")
	_ = os.MkdirAll(chunkDir, 0755)

	return &service{
		repo:         repo,
		logger:       logger,
		uploadDest:   cfg.Dest,
		maxSize:      cfg.MaxSize,
		allowedTypes: allowed,
		chunkDir:     chunkDir,
		ttlDays:      cfg.TTLDays,
	}
}

// Upload 上传单个文件
func (s *service) Upload(file *multipart.FileHeader) (*UploadResponse, error) {
	// 检查文件大小
	if file.Size > s.maxSize {
		return nil, fiber.NewError(fiber.StatusBadRequest,
			fmt.Sprintf("文件大小超过限制（最大 %d MB）", s.maxSize/1024/1024))
	}

	// 打开文件读取内容
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	// 通过 Magic Number 检测真实 MIME 类型
	buf := make([]byte, 512)
	n, _ := src.Read(buf)
	detectedType := http.DetectContentType(buf[:n])

	// 验证 MIME 类型
	if !s.allowedTypes[detectedType] {
		return nil, fiber.NewError(fiber.StatusBadRequest,
			fmt.Sprintf("不支持的文件类型: %s", detectedType))
	}

	// 重置文件读取位置
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("重置文件读取位置失败: %w", err)
	}

	// 生成存储文件名
	ext := filepath.Ext(file.Filename)
	storedName := uuid.New().String() + ext
	storedPath := filepath.Join(s.uploadDest, storedName)

	// 保存文件
	dst, err := os.Create(storedPath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		_ = os.Remove(storedPath) // 清理残留文件
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 创建数据库记录
	record := &Upload{
		Filename:     storedName,
		OriginalName: file.Filename,
		Mimetype:     detectedType,
		Size:         file.Size,
		Path:         storedPath,
	}

	if err := s.repo.Create(record); err != nil {
		_ = os.Remove(storedPath)
		return nil, fmt.Errorf("保存上传记录失败: %w", err)
	}

	s.logger.Info("文件上传成功",
		zap.String("uploadId", record.ID),
		zap.String("filename", storedName),
		zap.Int64("size", file.Size),
	)

	return ToUploadResponse(record), nil
}

// UploadChunk 上传单个分片
func (s *service) UploadChunk(req *ChunkUploadRequest, file *multipart.FileHeader) error {
	// 验证分片大小
	if file.Size > s.maxSize {
		return fiber.NewError(fiber.StatusBadRequest, "分片大小超过限制")
	}

	// 创建文件专属的分片目录
	fileChunkDir := filepath.Join(s.chunkDir, req.FileID)
	if err := os.MkdirAll(fileChunkDir, 0755); err != nil {
		return fmt.Errorf("创建分片目录失败: %w", err)
	}

	// 打开上传的分片
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("打开分片文件失败: %w", err)
	}
	defer src.Close()

	// 保存分片
	chunkPath := filepath.Join(fileChunkDir, fmt.Sprintf("chunk_%06d", req.ChunkIndex))
	dst, err := os.Create(chunkPath)
	if err != nil {
		return fmt.Errorf("创建分片文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("保存分片失败: %w", err)
	}

	s.logger.Debug("分片上传成功",
		zap.String("fileId", req.FileID),
		zap.Int("chunkIndex", req.ChunkIndex),
		zap.Int("totalChunks", req.TotalChunks),
	)

	return nil
}

// MergeChunks 合并分片
func (s *service) MergeChunks(req *MergeRequest) (*UploadResponse, error) {
	fileChunkDir := filepath.Join(s.chunkDir, req.FileID)

	// 检查分片目录是否存在
	if _, err := os.Stat(fileChunkDir); os.IsNotExist(err) {
		return nil, fiber.NewError(fiber.StatusBadRequest, "分片数据不存在")
	}

	// 生成存储文件名
	ext := filepath.Ext(req.Filename)
	storedName := uuid.New().String() + ext
	storedPath := filepath.Join(s.uploadDest, storedName)

	// 创建目标文件
	dst, err := os.Create(storedPath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	var totalSize int64

	// 按序号合并所有分片
	for i := 0; i < req.TotalChunks; i++ {
		chunkPath := filepath.Join(fileChunkDir, fmt.Sprintf("chunk_%06d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			_ = os.Remove(storedPath)
			return nil, fiber.NewError(fiber.StatusBadRequest,
				fmt.Sprintf("分片 %d 缺失", i))
		}

		n, err := io.Copy(dst, chunkFile)
		chunkFile.Close()
		if err != nil {
			_ = os.Remove(storedPath)
			return nil, fmt.Errorf("合并分片 %d 失败: %w", i, err)
		}
		totalSize += n
	}
	dst.Close()

	// 检测合并后文件的 MIME 类型
	detectedType, err := detectFileMIME(storedPath)
	if err != nil {
		_ = os.Remove(storedPath)
		return nil, fmt.Errorf("检测文件类型失败: %w", err)
	}

	if !s.allowedTypes[detectedType] {
		_ = os.Remove(storedPath)
		return nil, fiber.NewError(fiber.StatusBadRequest,
			fmt.Sprintf("不支持的文件类型: %s", detectedType))
	}

	// 检查总大小
	if totalSize > s.maxSize {
		_ = os.Remove(storedPath)
		return nil, fiber.NewError(fiber.StatusBadRequest,
			fmt.Sprintf("文件总大小超过限制（最大 %d MB）", s.maxSize/1024/1024))
	}

	// 创建数据库记录
	record := &Upload{
		Filename:     storedName,
		OriginalName: req.Filename,
		Mimetype:     detectedType,
		Size:         totalSize,
		Path:         storedPath,
	}

	if err := s.repo.Create(record); err != nil {
		_ = os.Remove(storedPath)
		return nil, fmt.Errorf("保存上传记录失败: %w", err)
	}

	// 清理分片目录
	_ = os.RemoveAll(fileChunkDir)

	s.logger.Info("分片合并成功",
		zap.String("uploadId", record.ID),
		zap.String("filename", storedName),
		zap.Int64("size", totalSize),
	)

	return ToUploadResponse(record), nil
}

// GetByID 按 ID 查询上传记录
func (s *service) GetByID(id string) (*UploadResponse, error) {
	upload, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "上传记录不存在")
	}
	return ToUploadResponse(upload), nil
}

// List 分页查询上传记录
func (s *service) List(page pagination.Pagination) ([]UploadResponse, int64, error) {
	uploads, total, err := s.repo.List(page)
	if err != nil {
		return nil, 0, fmt.Errorf("查询上传列表失败: %w", err)
	}

	responses := make([]UploadResponse, len(uploads))
	for i, u := range uploads {
		responses[i] = *ToUploadResponse(&u)
	}

	return responses, total, nil
}

// Delete 删除上传记录及磁盘文件
func (s *service) Delete(id string) error {
	upload, err := s.repo.FindByID(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "上传记录不存在")
	}

	// 删除磁盘文件
	if err := os.Remove(upload.Path); err != nil && !os.IsNotExist(err) {
		s.logger.Warn("删除磁盘文件失败",
			zap.String("path", upload.Path),
			zap.Error(err),
		)
	}

	// 删除数据库记录
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("删除上传记录失败: %w", err)
	}

	s.logger.Info("上传记录删除成功",
		zap.String("uploadId", id),
		zap.String("filename", upload.Filename),
	)

	return nil
}

// CleanExpired 清理过期上传文件
func (s *service) CleanExpired() error {
	cutoff := time.Now().AddDate(0, 0, -s.ttlDays)

	uploads, err := s.repo.DeleteExpired(cutoff)
	if err != nil {
		return fmt.Errorf("清理过期上传记录失败: %w", err)
	}

	if len(uploads) == 0 {
		return nil
	}

	// 删除对应的磁盘文件
	cleaned := 0
	for _, u := range uploads {
		if err := os.Remove(u.Path); err != nil && !os.IsNotExist(err) {
			s.logger.Warn("清理过期文件失败",
				zap.String("path", u.Path),
				zap.Error(err),
			)
			continue
		}
		cleaned++
	}

	s.logger.Info("过期文件清理完成",
		zap.Int("total", len(uploads)),
		zap.Int("cleaned", cleaned),
	)

	return nil
}

// detectFileMIME 通过读取文件头部检测 MIME 类型
func detectFileMIME(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(buf[:n]), nil
}
