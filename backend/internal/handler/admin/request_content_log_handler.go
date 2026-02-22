package admin

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// RequestContentLogHandler 请求内容日志管理 Handler
type RequestContentLogHandler struct {
	svc *service.RequestContentLogService
}

// NewRequestContentLogHandler 创建请求内容日志 Handler
func NewRequestContentLogHandler(svc *service.RequestContentLogService) *RequestContentLogHandler {
	return &RequestContentLogHandler{svc: svc}
}

// List 分页查询请求内容日志
// GET /api/v1/admin/request-content-logs
func (h *RequestContentLogHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)

	filters := service.RequestContentLogFilters{}

	if v := c.Query("user_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filters.UserID = id
		}
	}
	if v := c.Query("api_key_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filters.APIKeyID = id
		}
	}
	if v := c.Query("model"); v != "" {
		filters.Model = v
	}
	if v := c.Query("platform"); v != "" {
		filters.Platform = v
	}
	if v := c.Query("start_date"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filters.StartDate = t
		}
	}
	if v := c.Query("end_date"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filters.EndDate = t
		}
	}

	logs, total, err := h.svc.List(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, logs, total, page, pageSize)
}

// GetByID 查询请求内容日志详情
// GET /api/v1/admin/request-content-logs/:id
func (h *RequestContentLogHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	log, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, log)
}
