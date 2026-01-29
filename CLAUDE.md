# Sub2API 项目说明

## 钉钉登录自动分配分组

钉钉登录创建的新用户会自动分配以下三个分组：

- Claude code 每日50美刀分组
- codex 每日50美刀分组
- Gemini 每日50美刀分组

**注意**：这些分组必须在系统中预先创建，否则自动分配会静默失败（仅记录日志）。

### 相关代码位置

- `backend/internal/service/auth_service.go` - `DingTalkDefaultGroupNames` 变量定义默认分组名称
- `backend/internal/service/auth_service.go` - `assignDingTalkDefaultGroups` 方法实现分配逻辑

### 修改默认分组

如需修改默认分配的分组，编辑 `auth_service.go` 中的 `DingTalkDefaultGroupNames` 变量：

```go
var DingTalkDefaultGroupNames = []string{
    "Claude code 每日50美刀分组",
    "codex 每日50美刀分组",
    "Gemini 每日50美刀分组",
}
```

## 部署流程

1. 提交并推送代码到 GitHub：
   ```bash
   git add . && git commit -m "your message"
   export HOME=/root && gh auth setup-git
   git push myfork main
   ```

2. 创建/更新 tag（替换 `x` 为版本号）：
   ```bash
   git tag -d v1.0.x-dingtalk 2>/dev/null || true
   git push myfork :refs/tags/v1.0.x-dingtalk 2>/dev/null || true
   git tag v1.0.x-dingtalk
   git push myfork v1.0.x-dingtalk
   ```

3. 触发 GitHub Actions 打包：
   ```bash
   gh workflow run Release --repo wangxiaobo775/sub2api --ref v1.0.x-dingtalk -f tag=v1.0.x-dingtalk -f simple_release=true
   ```

4. 监控打包进度：
   ```bash
   gh run list --repo wangxiaobo775/sub2api --workflow=Release --limit 3
   gh run watch <run_id> --repo wangxiaobo775/sub2api
   ```

5. 更新 `deploy/docker-compose.yml` 中的镜像版本

6. 部署：
   ```bash
   cd /opt/nezha/agent/sub2api/deploy && docker compose pull sub2api && docker compose up -d sub2api
   ```

7. 检查日志：
   ```bash
   docker logs sub2api --tail 20
   ```
