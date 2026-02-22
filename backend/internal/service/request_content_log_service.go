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

// sessionTracker 会话追踪器：同一用户+API Key 在时间窗口内的请求归为同一会话
type sessionTracker struct {
	fingerprint string
	lastSeen    time.Time
}

// sessionWindow 会话超时窗口：超过此时间视为新会话
const sessionWindow = 30 * time.Minute

// RequestContentLogService 请求内容日志服务
type RequestContentLogService struct {
	repo     RequestContentLogRepository
	cfg      *config.RequestContentLogConfig
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup

	// 会话追踪缓存: "userID:apiKeyID" → sessionTracker
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

// MaxSize 返回单条 messages 最大存储字节（完整存储模式默认 512KB）
func (s *RequestContentLogService) MaxSize() int {
	if s.cfg == nil || s.cfg.MaxSize <= 0 {
		return 524288
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

// LogAsync 异步提取 messages 并存储完整对话（在 goroutine 中调用）
// 每次请求存储完整的 messages 数组，非文本内容（图片、tool_use 等）精简为占位符以节省空间
func (s *RequestContentLogService) LogAsync(body []byte, userID, apiKeyID int64, clientIP, ua, platform string) {
	if !s.IsEnabled() {
		return
	}

	// 部分解析 JSON，提取 model + messages/contents
	var partial struct {
		Model    string          `json:"model"`
		Messages json.RawMessage `json:"messages"`
		Contents json.RawMessage `json:"contents"` // Gemini 格式
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

	totalCount := len(messageArray)

	// 基于时间窗口的会话识别：同一 userID+apiKeyID 在 30 分钟内的请求归为同一会话
	fingerprint := s.resolveSessionFingerprint(userID, apiKeyID)

	// 精简消息内容：保留全部文本，移除大体积非文本内容（图片、base64 等）
	simplified := simplifyMessages(messageArray, platform)

	// 序列化
	messagesJSON, err := json.Marshal(simplified)
	if err != nil {
		return
	}

	// 截断到最大大小
	maxSize := s.MaxSize()
	if len(messagesJSON) > maxSize {
		messagesJSON = messagesJSON[:maxSize]
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
		Messages:           messagesJSON,
		Platform:           platform,
		IPAddress:          clientIP,
		UserAgent:          ua,
		SessionFingerprint: fingerprint,
		MessageOffset:      0,
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

// resolveSessionFingerprint 基于时间窗口的会话识别
// 同一 userID+apiKeyID 在 sessionWindow 内的请求归为同一会话
// 超时后自动生成新指纹（新会话）
func (s *RequestContentLogService) resolveSessionFingerprint(userID, apiKeyID int64) string {
	cacheKey := fmt.Sprintf("%d:%d", userID, apiKeyID)
	now := time.Now()

	if val, ok := s.sessionCache.Load(cacheKey); ok {
		tracker := val.(*sessionTracker)
		if now.Sub(tracker.lastSeen) < sessionWindow {
			// 在时间窗口内：复用同一指纹，更新最后活跃时间
			tracker.lastSeen = now
			return tracker.fingerprint
		}
	}

	// 超时或新用户：生成新指纹
	raw := fmt.Sprintf("%d:%d:%d", userID, apiKeyID, now.UnixNano())
	hash := sha256.Sum256([]byte(raw))
	fp := hex.EncodeToString(hash[:])[:16]

	s.sessionCache.Store(cacheKey, &sessionTracker{
		fingerprint: fp,
		lastSeen:    now,
	})
	return fp
}

// evictStaleSessions 清理超过 1 小时未活跃的会话追踪缓存
func (s *RequestContentLogService) evictStaleSessions() {
	threshold := time.Now().Add(-1 * time.Hour)
	s.sessionCache.Range(func(key, value any) bool {
		tracker := value.(*sessionTracker)
		if tracker.lastSeen.Before(threshold) {
			s.sessionCache.Delete(key)
		}
		return true
	})
}

// simplifyMessages 精简消息数组：保留所有文本内容，移除大体积非文本内容
// - user 消息：完整保留文本；图片/文件替换为占位符
// - assistant 消息：保留文本；tool_use 仅保留名称和输入摘要
// - system 消息：完整保留
func simplifyMessages(messages []json.RawMessage, _ string) []json.RawMessage {
	result := make([]json.RawMessage, 0, len(messages))

	for _, msg := range messages {
		simplified := simplifyOneMessage(msg)
		if simplified != nil {
			result = append(result, simplified)
		}
	}
	return result
}

// simplifyOneMessage 精简单条消息
func simplifyOneMessage(raw json.RawMessage) json.RawMessage {
	// 先解析出基本结构
	var base map[string]json.RawMessage
	if err := json.Unmarshal(raw, &base); err != nil {
		return raw // 解析失败则原样返回
	}

	// Gemini 格式：精简 parts
	if partsRaw, ok := base["parts"]; ok {
		base["parts"] = simplifyGeminiParts(partsRaw)
		out, _ := json.Marshal(base)
		return out
	}

	// OpenAI/Anthropic 格式：精简 content
	contentRaw, hasContent := base["content"]
	if !hasContent {
		// 没有 content 字段（可能只有 role），原样返回
		out, _ := json.Marshal(base)
		return out
	}

	// content 是字符串 → 完整保留（OpenAI 格式）
	if contentRaw[0] == '"' {
		out, _ := json.Marshal(base)
		return out
	}

	// content 是数组（Anthropic content blocks）→ 精简
	var blocks []map[string]any
	if err := json.Unmarshal(contentRaw, &blocks); err == nil {
		simplified := simplifyContentBlocks(blocks)
		simplifiedJSON, _ := json.Marshal(simplified)
		base["content"] = simplifiedJSON
		out, _ := json.Marshal(base)
		return out
	}

	// 其他格式，原样返回
	out, _ := json.Marshal(base)
	return out
}

// simplifyContentBlocks 精简 Anthropic content blocks
// 保留 text 块完整内容，其他块（image/tool_use/tool_result）替换为摘要
func simplifyContentBlocks(blocks []map[string]any) []map[string]any {
	result := make([]map[string]any, 0, len(blocks))
	for _, block := range blocks {
		blockType, _ := block["type"].(string)
		switch blockType {
		case "text":
			// 文本块完整保留
			result = append(result, block)
		case "image":
			// 图片替换为占位符
			result = append(result, map[string]any{
				"type": "text",
				"text": "[image]",
			})
		case "tool_use":
			// 工具调用：保留名称，省略输入详情
			name, _ := block["name"].(string)
			result = append(result, map[string]any{
				"type": "text",
				"text": fmt.Sprintf("[tool_use: %s]", name),
			})
		case "tool_result":
			// 工具结果：仅保留占位符
			result = append(result, map[string]any{
				"type": "text",
				"text": "[tool_result]",
			})
		default:
			// 未知类型：保留 type 标记
			result = append(result, map[string]any{
				"type": "text",
				"text": fmt.Sprintf("[%s]", blockType),
			})
		}
	}
	return result
}

// simplifyGeminiParts 精简 Gemini parts
// 保留 text 部分，移除 inline_data（图片等大体积内容）
func simplifyGeminiParts(raw json.RawMessage) json.RawMessage {
	var parts []map[string]any
	if err := json.Unmarshal(raw, &parts); err != nil {
		return raw
	}

	simplified := make([]map[string]any, 0, len(parts))
	for _, part := range parts {
		if _, hasText := part["text"]; hasText {
			// 文本部分完整保留
			simplified = append(simplified, part)
		} else if _, hasInline := part["inline_data"]; hasInline {
			// 内联数据（图片等）替换为占位符
			simplified = append(simplified, map[string]any{"text": "[inline_data]"})
		} else if _, hasFuncCall := part["functionCall"]; hasFuncCall {
			// 函数调用替换为占位符
			simplified = append(simplified, map[string]any{"text": "[functionCall]"})
		} else if _, hasFuncResp := part["functionResponse"]; hasFuncResp {
			simplified = append(simplified, map[string]any{"text": "[functionResponse]"})
		} else {
			// 未知类型原样保留
			simplified = append(simplified, part)
		}
	}

	out, _ := json.Marshal(simplified)
	return out
}
