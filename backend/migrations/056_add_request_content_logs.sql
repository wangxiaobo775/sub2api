-- Request Content Logs table
-- 记录 API 请求中的 messages 内容，供管理员事后审计

CREATE TABLE IF NOT EXISTS request_content_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    api_key_id BIGINT,
    model VARCHAR(100),
    messages JSONB,
    platform VARCHAR(20),
    ip_address VARCHAR(45),
    user_agent VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引：按用户查询
CREATE INDEX IF NOT EXISTS idx_request_content_logs_user_id
    ON request_content_logs (user_id);

-- 索引：按 API 密钥查询
CREATE INDEX IF NOT EXISTS idx_request_content_logs_api_key_id
    ON request_content_logs (api_key_id);

-- 索引：按创建时间查询（也用于清理任务）
CREATE INDEX IF NOT EXISTS idx_request_content_logs_created_at
    ON request_content_logs (created_at);

-- 索引：按模型查询
CREATE INDEX IF NOT EXISTS idx_request_content_logs_model
    ON request_content_logs (model);

-- 复合索引：按用户+时间范围查询
CREATE INDEX IF NOT EXISTS idx_request_content_logs_user_created
    ON request_content_logs (user_id, created_at);
