package repository

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type requestContentLogRepository struct {
	db *sql.DB
}

// NewRequestContentLogRepository 创建请求内容日志仓储
func NewRequestContentLogRepository(sqlDB *sql.DB) service.RequestContentLogRepository {
	return &requestContentLogRepository{db: sqlDB}
}

func (r *requestContentLogRepository) Create(ctx context.Context, log *service.RequestContentLog) error {
	query := `
		INSERT INTO request_content_logs (user_id, api_key_id, model, messages, platform, ip_address, user_agent,
		                                  session_fingerprint, message_offset, message_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		log.UserID,
		log.APIKeyID,
		log.Model,
		log.Messages,
		log.Platform,
		log.IPAddress,
		log.UserAgent,
		nullableString(log.SessionFingerprint),
		log.MessageOffset,
		log.MessageCount,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *requestContentLogRepository) List(ctx context.Context, filters service.RequestContentLogFilters, page, pageSize int) ([]*service.RequestContentLog, int64, error) {
	// 构建 WHERE 子句
	where := "WHERE 1=1"
	args := make([]any, 0)
	argIdx := 1

	if filters.UserID > 0 {
		where += " AND rcl.user_id = $" + strconv.Itoa(argIdx)
		args = append(args, filters.UserID)
		argIdx++
	}
	if filters.APIKeyID > 0 {
		where += " AND rcl.api_key_id = $" + strconv.Itoa(argIdx)
		args = append(args, filters.APIKeyID)
		argIdx++
	}
	if filters.Model != "" {
		where += " AND rcl.model = $" + strconv.Itoa(argIdx)
		args = append(args, filters.Model)
		argIdx++
	}
	if filters.Platform != "" {
		where += " AND rcl.platform = $" + strconv.Itoa(argIdx)
		args = append(args, filters.Platform)
		argIdx++
	}
	if filters.SessionFingerprint != "" {
		where += " AND rcl.session_fingerprint = $" + strconv.Itoa(argIdx)
		args = append(args, filters.SessionFingerprint)
		argIdx++
	}
	if !filters.StartDate.IsZero() {
		where += " AND rcl.created_at >= $" + strconv.Itoa(argIdx)
		args = append(args, filters.StartDate)
		argIdx++
	}
	if !filters.EndDate.IsZero() {
		where += " AND rcl.created_at < $" + strconv.Itoa(argIdx)
		args = append(args, filters.EndDate)
		argIdx++
	}

	// 计算总数
	countQuery := "SELECT COUNT(*) FROM request_content_logs rcl " + where
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*service.RequestContentLog{}, 0, nil
	}

	// 查询列表（不返回 messages 内容以减少传输量）
	listQuery := `
		SELECT rcl.id, rcl.user_id, rcl.api_key_id, rcl.model, rcl.platform,
		       rcl.ip_address, rcl.user_agent, rcl.created_at,
		       rcl.session_fingerprint, rcl.message_offset, rcl.message_count,
		       u.email AS user_email,
		       ak.name AS api_key_name
		FROM request_content_logs rcl
		LEFT JOIN users u ON rcl.user_id = u.id
		LEFT JOIN api_keys ak ON rcl.api_key_id = ak.id
		` + where + `
		ORDER BY rcl.created_at DESC
		LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var logs []*service.RequestContentLog
	for rows.Next() {
		log := &service.RequestContentLog{}
		var userEmail, apiKeyName, sessionFP sql.NullString
		if err := rows.Scan(
			&log.ID, &log.UserID, &log.APIKeyID, &log.Model, &log.Platform,
			&log.IPAddress, &log.UserAgent, &log.CreatedAt,
			&sessionFP, &log.MessageOffset, &log.MessageCount,
			&userEmail, &apiKeyName,
		); err != nil {
			return nil, 0, err
		}
		log.UserEmail = userEmail.String
		log.APIKeyName = apiKeyName.String
		log.SessionFingerprint = sessionFP.String
		logs = append(logs, log)
	}

	return logs, total, rows.Err()
}

func (r *requestContentLogRepository) GetByID(ctx context.Context, id int64) (*service.RequestContentLog, error) {
	query := `
		SELECT rcl.id, rcl.user_id, rcl.api_key_id, rcl.model, rcl.messages, rcl.platform,
		       rcl.ip_address, rcl.user_agent, rcl.created_at,
		       rcl.session_fingerprint, rcl.message_offset, rcl.message_count,
		       u.email AS user_email,
		       ak.name AS api_key_name
		FROM request_content_logs rcl
		LEFT JOIN users u ON rcl.user_id = u.id
		LEFT JOIN api_keys ak ON rcl.api_key_id = ak.id
		WHERE rcl.id = $1
	`
	log := &service.RequestContentLog{}
	var userEmail, apiKeyName, sessionFP sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID, &log.UserID, &log.APIKeyID, &log.Model, &log.Messages, &log.Platform,
		&log.IPAddress, &log.UserAgent, &log.CreatedAt,
		&sessionFP, &log.MessageOffset, &log.MessageCount,
		&userEmail, &apiKeyName,
	)
	if err != nil {
		return nil, err
	}
	log.UserEmail = userEmail.String
	log.APIKeyName = apiKeyName.String
	log.SessionFingerprint = sessionFP.String
	return log, nil
}

func (r *requestContentLogRepository) ListBySession(ctx context.Context, fingerprint string) ([]*service.RequestContentLog, error) {
	query := `
		SELECT rcl.id, rcl.user_id, rcl.api_key_id, rcl.model, rcl.messages, rcl.platform,
		       rcl.ip_address, rcl.user_agent, rcl.created_at,
		       rcl.session_fingerprint, rcl.message_offset, rcl.message_count,
		       u.email AS user_email,
		       ak.name AS api_key_name
		FROM request_content_logs rcl
		LEFT JOIN users u ON rcl.user_id = u.id
		LEFT JOIN api_keys ak ON rcl.api_key_id = ak.id
		WHERE rcl.session_fingerprint = $1
		ORDER BY rcl.created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, fingerprint)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var logs []*service.RequestContentLog
	for rows.Next() {
		log := &service.RequestContentLog{}
		var userEmail, apiKeyName, sessionFP sql.NullString
		if err := rows.Scan(
			&log.ID, &log.UserID, &log.APIKeyID, &log.Model, &log.Messages, &log.Platform,
			&log.IPAddress, &log.UserAgent, &log.CreatedAt,
			&sessionFP, &log.MessageOffset, &log.MessageCount,
			&userEmail, &apiKeyName,
		); err != nil {
			return nil, err
		}
		log.UserEmail = userEmail.String
		log.APIKeyName = apiKeyName.String
		log.SessionFingerprint = sessionFP.String
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

func (r *requestContentLogRepository) DeleteBefore(ctx context.Context, retentionDays int) (int64, error) {
	query := `DELETE FROM request_content_logs WHERE created_at < NOW() - MAKE_INTERVAL(days => $1)`
	result, err := r.db.ExecContext(ctx, query, retentionDays)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// nullableString 将空字符串转为 sql.NullString
func nullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
