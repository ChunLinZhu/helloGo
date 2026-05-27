// Package upload — 上传请求/响应结构体
package upload

// UploadResponse 上传响应
type UploadResponse struct {
	ID           string `json:"id"`
	Filename     string `json:"filename"`
	OriginalName string `json:"originalName"`
	Mimetype     string `json:"mimetype"`
	Size         int64  `json:"size"`
	Path         string `json:"path"`
	URL          string `json:"url"` // 可访问的 URL 路径
}

// ToUploadResponse 将 Upload 模型转换为响应结构
func ToUploadResponse(u *Upload) *UploadResponse {
	return &UploadResponse{
		ID:           u.ID,
		Filename:     u.Filename,
		OriginalName: u.OriginalName,
		Mimetype:     u.Mimetype,
		Size:         u.Size,
		Path:         u.Path,
		URL:          "/uploads/" + u.Filename,
	}
}

// ChunkUploadRequest 分片上传请求
type ChunkUploadRequest struct {
	FileID      string `form:"fileId" validate:"required"`      // 文件唯一标识（客户端生成）
	ChunkIndex  int    `form:"chunkIndex" validate:"gte=0"`     // 分片序号（从 0 开始）
	TotalChunks int    `form:"totalChunks" validate:"required,gte=1"` // 总分片数
	Filename    string `form:"filename" validate:"required"`    // 原始文件名
}

// MergeRequest 合并分片请求
type MergeRequest struct {
	FileID      string `json:"fileId" validate:"required"`      // 文件唯一标识
	Filename    string `json:"filename" validate:"required"`    // 原始文件名
	TotalChunks int    `json:"totalChunks" validate:"required,gte=1"` // 总分片数
}
