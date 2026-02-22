package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

	// Session 增量存储字段
	SessionFingerprint string `json:"session_fingerprint,omitempty"`
	MessageOffset      int    `json:"message_offset"`
	MessageCount       int    `json:"message_count"`

	// JOIN 字段（仅列表/详情查询时填充）
	UserEmail  string `json:"user_email,omitempty"`
	APIKeyName string `json:"api_key_name,omitempty"`
}

// RequestContentLogFilters 列表查询过滤条件
type RequestContentLogFilters struct {
	UserID             int64
	APIKeyID           int64
	Model              string
	Platform           string
	SessionFingerprint string
	StartDate          time.Time
	EndDate            time.Time
}

// RequestContentLogRepository 请求内容日志仓储接口
type RequestContentLogRepository interface {
	Create(ctx context.Context, log *RequestContentLog) error
	List(ctx context.Context, filters RequestContentLogFilters, page, pageSize int) ([]*RequestContentLog, int64, error)
	GetByID(ctx context.Context, id int64) (*RequestContentLog, error)
	ListBySession(ctx context.Context, fingerprint string) ([]*RequestContentLog, error)
	DeleteBefore(ctx context.Context, retentionDays int) (int64, error)
}

// sessionState 内存中的 session 状态（LRU 缓存条目）
type sessionState struct {
	messageCount int
	lastSeen     time.Time
}

// RequestContentLogService 请求内容日志服务
type RequestContentLogService struct {
	repo     RequestContentLogRepository
	cfg      *config.RequestContentLogConfig
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup

	// session LRU 缓存: fingerprint → sessionState
	sessionCache sync.Map
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

	// 启动清理 goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.cleanup()
				s.evictStaleSessions()
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

// LogAsync 异步提取 messages 并增量存储（在 goroutine 中调用）
func (s *RequestContentLogService) LogAsync(body []byte, userID, apiKeyID int64, clientIP, ua, platform string) {
	if !s.IsEnabled() {
		return
	}

	// 部分解析 JSON，提取 model + messages/contents
	var partial struct {
		Model    string            `json:"model"`
		Messages json.RawMessage   `json:"messages"`
		Contents json.RawMessage   `json:"contents"` // Gemini 格式
	}
	if err := json.Unmarshal(body, &partial); err != nil {
		return
	}

	// 兼容 Gemini contents 格式
	rawMessages := partial.Messages
	if len(rawMessages) == 0 {
		rawMessages = partial.Contents
	}
	if len(rawMessages) == 0 {
		return
	}

	// 解析消息数组为独立元素
	var messageArray []json.RawMessage
	if err := json.Unmarshal(rawMessages, &messageArray); err != nil || len(messageArray) == 0 {
		return
	}

	// 提取第一条 user 消息的内容，用于计算 session fingerprint
	firstUserContent := extractFirstUserContent(messageArray, platform)

	// 计算 session fingerprint
	fingerprint := computeSessionFingerprint(userID, apiKeyID, firstUserContent)

	// 查 LRU 缓存获取上一次的 message_count
	totalCount := len(messageArray)
	var deltaMessages []json.RawMessage
	var messageOffset int

	if val, ok := s.sessionCache.Load(fingerprint); ok {
		state := val.(*sessionState)
		prevCount := state.messageCount
		if totalCount > prevCount {
			// 增量：只保存新增的消息
			deltaMessages = messageArray[prevCount:]
			messageOffset = prevCount
		} else {
			// 对话重置或消息数相同/减少，存储全部
			deltaMessages = messageArray
			messageOffset = 0
		}
	} else {
		// 新会话，存储全部
		deltaMessages = messageArray
		messageOffset = 0
	}

	// 更新缓存
	s.sessionCache.Store(fingerprint, &sessionState{
		messageCount: totalCount,
		lastSeen:     time.Now(),
	})

	// 将 delta messages 序列化
	deltaJSON, err := json.Marshal(deltaMessages)
	if err != nil {
		return
	}

	// 截断 messages 到最大大小
	maxSize := s.MaxSize()
	if len(deltaJSON) > maxSize {
		deltaJSON = deltaJSON[:maxSize]
	}

	// 截断 User-Agent
	if len(ua) > 512 {
		ua = ua[:512]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logEntry := &RequestContentLog{
		UserID:             userID,
		APIKeyID:           apiKeyID,
		Model:              partial.Model,
		Messages:           deltaJSON,
		Platform:           platform,
		IPAddress:          clientIP,
		UserAgent:          ua,
		SessionFingerprint: fingerprint,
		MessageOffset:      messageOffset,
		MessageCount:       totalCount,
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

// ListBySession 按会话查询完整对话流
func (s *RequestContentLogService) ListBySession(ctx context.Context, fingerprint string) ([]*RequestContentLog, error) {
	return s.repo.ListBySession(ctx, fingerprint)
}

// cleanup 执行一次过期数据清理
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

// evictStaleSessions 清理超过 1 小时未访问的 session 缓存条目
func (s *RequestContentLogService) evictStaleSessions() {
	threshold := time.Now().Add(-1 * time.Hour)
	s.sessionCache.Range(func(key, value any) bool {
		state := value.(*sessionState)
		if state.lastSeen.Before(threshold) {
			s.sessionCache.Delete(key)
		}
		return true
	})
}

// computeSessionFingerprint 计算 session 指纹
// SHA256(userID:apiKeyID:firstUserContent)[:16]
func computeSessionFingerprint(userID, apiKeyID int64, firstUserContent string) string {
	raw := fmt.Sprintf("%d:%d:%s", userID, apiKeyID, firstUserContent)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])[:16]
}

// extractFirstUserContent 从消息数组中提取第一条 user 角色消息的文本内容
// 兼容 OpenAI/Anthropic (role=user) 和 Gemini (role=user, parts) 格式
func extractFirstUserContent(messages []json.RawMessage, platform string) string {
	for _, msg := range messages {
		var parsed struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
			Parts   json.RawMessage `json:"parts"` // Gemini 格式
		}
		if err := json.Unmarshal(msg, &parsed); err != nil {
			continue
		}
		if parsed.Role != "user" {
			continue
		}

		// 尝试从 content 获取文本
		text := extractTextFromField(parsed.Content)
		if text != "" {
			return text
		}

		// Gemini 格式：从 parts 获取文本
		text = extractTextFromParts(parsed.Parts)
		if text != "" {
			return text
		}
	}
	return ""
}

// extractTextFromField 从 content 字段提取文本
// content 可能是字符串 "hello" 或数组 [{"type":"text","text":"hello"}, ...]
func extractTextFromField(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// 尝试作为字符串解析
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}

	// 尝试作为 content blocks 数组解析（Anthropic 格式）
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &blocks); err == nil {
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				return b.Text
			}
		}
	}

	return ""
}

// extractTextFromParts 从 Gemini parts 字段提取文本
// parts: [{"text": "hello"}, ...]
func extractTextFromParts(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var parts []struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		for _, p := range parts {
			if p.Text != "" {
				return p.Text
			}
		}
	}

	return ""
}
