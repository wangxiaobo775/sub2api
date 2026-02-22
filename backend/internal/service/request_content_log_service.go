package service

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// RequestContentLog 请求内容日志数据模型
type RequestContentLog struct {
	ID        int64           `json:"id"`
	UserID    int64           `json:"user_id"`
	APIKeyID  int64           `json:"api_key_id"`
	Model     string          `json:"model"`
	Messages  json.RawMessage `json:"messages,omitempty"`
	Platform  string          `json:"platform"`
	IPAddress string          `json:"ip_address"`
	UserAgent string          `json:"user_agent"`
	CreatedAt time.Time       `json:"created_at"`

	// JOIN 字段（仅列表/详情查询时填充）
	UserEmail  string `json:"user_email,omitempty"`
	APIKeyName string `json:"api_key_name,omitempty"`
}

// RequestContentLogFilters 列表查询过滤条件
type RequestContentLogFilters struct {
	UserID    int64
	APIKeyID  int64
	Model     string
	Platform  string
	StartDate time.Time
	EndDate   time.Time
}

// RequestContentLogRepository 请求内容日志仓储接口
type RequestContentLogRepository interface {
	Create(ctx context.Context, log *RequestContentLog) error
	List(ctx context.Context, filters RequestContentLogFilters, page, pageSize int) ([]*RequestContentLog, int64, error)
	GetByID(ctx context.Context, id int64) (*RequestContentLog, error)
	DeleteBefore(ctx context.Context, retentionDays int) (int64, error)
}

// RequestContentLogService 请求内容日志服务
type RequestContentLogService struct {
	repo     RequestContentLogRepository
	cfg      *config.RequestContentLogConfig
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

// NewRequestContentLogService 创建请求内容日志服务
func NewRequestContentLogService(repo RequestContentLogRepository, cfg *config.RequestContentLogConfig) *RequestContentLogService {
	return &RequestContentLogService{
		repo:   repo,
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
}

// ProvideRequestContentLogService 创建并启动请求内容日志服务
func ProvideRequestContentLogService(repo RequestContentLogRepository, cfg *config.Config) *RequestContentLogService {
	svc := NewRequestContentLogService(repo, &cfg.RequestContentLog)
	svc.Start()
	return svc
}

// IsEnabled 检查功能是否启用
func (s *RequestContentLogService) IsEnabled() bool {
	return s != nil && s.cfg != nil && s.cfg.Enabled
}

// MaxSize 返回单条 messages 最大存储字节
func (s *RequestContentLogService) MaxSize() int {
	if s.cfg == nil || s.cfg.MaxSize <= 0 {
		return 65536
	}
	return s.cfg.MaxSize
}

// Start 启动后台清理任务
func (s *RequestContentLogService) Start() {
	if s == nil || !s.IsEnabled() {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.cleanup()
			case <-s.stopCh:
				return
			}
		}
	}()
	log.Printf("[RequestContentLog] Cleanup goroutine started (retention=%d days)", s.cfg.RetentionDays)
}

// Stop 优雅停止
func (s *RequestContentLogService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

// LogAsync 异步提取 messages 并存储（在 goroutine 中调用）
func (s *RequestContentLogService) LogAsync(body []byte, userID, apiKeyID int64, clientIP, ua, platform string) {
	if !s.IsEnabled() {
		return
	}

	// 部分解析 JSON，仅提取 model + messages
	var partial struct {
		Model    string          `json:"model"`
		Messages json.RawMessage `json:"messages"`
	}
	if err := json.Unmarshal(body, &partial); err != nil || len(partial.Messages) == 0 {
		return
	}

	// 截断 messages 到最大大小
	messages := partial.Messages
	maxSize := s.MaxSize()
	if len(messages) > maxSize {
		messages = messages[:maxSize]
	}

	// 截断 User-Agent
	if len(ua) > 512 {
		ua = ua[:512]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logEntry := &RequestContentLog{
		UserID:    userID,
		APIKeyID:  apiKeyID,
		Model:     partial.Model,
		Messages:  messages,
		Platform:  platform,
		IPAddress: clientIP,
		UserAgent: ua,
	}

	if err := s.repo.Create(ctx, logEntry); err != nil {
		log.Printf("[RequestContentLog] Failed to save: %v", err)
	}
}

// List 查询列表
func (s *RequestContentLogService) List(ctx context.Context, filters RequestContentLogFilters, page, pageSize int) ([]*RequestContentLog, int64, error) {
	return s.repo.List(ctx, filters, page, pageSize)
}

// GetByID 查询详情
func (s *RequestContentLogService) GetByID(ctx context.Context, id int64) (*RequestContentLog, error) {
	return s.repo.GetByID(ctx, id)
}

// cleanup 执行一次清理
func (s *RequestContentLogService) cleanup() {
	retentionDays := 30
	if s.cfg != nil && s.cfg.RetentionDays > 0 {
		retentionDays = s.cfg.RetentionDays
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	deleted, err := s.repo.DeleteBefore(ctx, retentionDays)
	if err != nil {
		log.Printf("[RequestContentLog] Cleanup failed: %v", err)
		return
	}
	if deleted > 0 {
		log.Printf("[RequestContentLog] Cleaned up %d records older than %d days", deleted, retentionDays)
	}
}
