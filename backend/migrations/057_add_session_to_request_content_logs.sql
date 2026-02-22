-- 请求内容日志增量存储（Session 去重）
-- 新增 session 指纹、消息偏移量、消息总数字段

ALTER TABLE request_content_logs
    ADD COLUMN IF NOT EXISTS session_fingerprint VARCHAR(16),
    ADD COLUMN IF NOT EXISTS message_offset INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS message_count INT NOT NULL DEFAULT 0;

-- 索引：按 session 查询完整对话流
CREATE INDEX IF NOT EXISTS idx_rcl_session
    ON request_content_logs (session_fingerprint, created_at);
