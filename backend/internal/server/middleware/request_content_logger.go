package middleware

import (
	"bytes"
	"io"
	"log"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// bodyCapture 透明包装 io.ReadCloser，在 handler 读取时同步复制到 buffer
type bodyCapture struct {
	io.ReadCloser
	buf bytes.Buffer
}

func (bc *bodyCapture) Read(p []byte) (int, error) {
	n, err := bc.ReadCloser.Read(p)
	if n > 0 {
		_, _ = bc.buf.Write(p[:n])
	}
	return n, err
}

// RequestContentLogger 请求内容记录中间件
// 使用 TeeReader 模式透明捕获请求体，handler 正常读取不受影响
func RequestContentLogger(svc *service.RequestContentLogService, platform string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "POST" || svc == nil || !svc.IsEnabled() {
			c.Next()
			return
		}

		if c.Request.Body == nil {
			c.Next()
			return
		}

		// 捕获请求体：不设上限，上游 RequestBodyLimit 已限制总大小
		capture := &bodyCapture{
			ReadCloser: c.Request.Body,
		}
		c.Request.Body = capture

		c.Next()

		body := capture.buf.Bytes()
		if len(body) == 0 || c.Writer.Status() >= 500 {
			return
		}

		// 从 context 获取已认证的用户信息（apiKeyAuth 中间件已设置）
		var userID, apiKeyID int64
		if apiKey, ok := GetAPIKeyFromContext(c); ok && apiKey != nil {
			apiKeyID = apiKey.ID
			if apiKey.User != nil {
				userID = apiKey.User.ID
			}
		}

		clientIP := ip.GetClientIP(c)
		ua := c.GetHeader("User-Agent")

		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[RequestContentLog] Panic in async log: %v", r)
				}
			}()
			svc.LogAsync(body, userID, apiKeyID, clientIP, ua, platform)
		}()
	}
}
