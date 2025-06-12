package dto

// ProcessVideoRequest 处理视频请求DTO
type ProcessVideoRequest struct {
	InputPath string `json:"inputPath"`
}

// ProcessVideoResponse 处理视频响应DTO
type ProcessVideoResponse struct {
	Success        bool     `json:"success"`
	ProcessedFiles []string `json:"processedFiles"`
	ErrorMessage   string   `json:"errorMessage,omitempty"`
}

// BatchProcessRequest 批量处理请求DTO
type BatchProcessRequest struct {
	InputDir   string `json:"inputDir"`
	OutputDir  string `json:"outputDir"`
	ConfigPath string `json:"configPath,omitempty"`
}

// BatchProcessResponse 批量处理响应DTO
type BatchProcessResponse struct {
	TotalVideos     int      `json:"totalVideos"`
	ProcessedVideos int      `json:"processedVideos"`
	FailedVideos    int      `json:"failedVideos"`
	FailedFiles     []string `json:"failedFiles,omitempty"`
	ErrorMessage    string   `json:"errorMessage,omitempty"`
}
