# 后端部署

`Dockerfile.backend` 面向 API 文档建议的 Go 后端目录：`backend/go.mod` 与 `backend/cmd/server/main.go`。后端代码合入后，使用以下命令构建并启动：

```powershell
docker compose --env-file .env -f infra/docker/docker-compose.backend.yml up --build -d
```

服务默认绑定宿主机 `8080`；可在环境中设置 `API_PORT` 覆盖。生产环境请使用真实的 `LLM_API_KEY`、`OCR_API_KEY`、`DATABASE_URL`，不要提交 `.env`。当前的 `docker-compose.yml` 保留给前端开发容器，两者可并存。

部署后建议依次执行 fixture 校验、API 契约测试与 k6 冒烟压测：

```powershell
python backend/tests/verify_fixtures.py
$env:API_BASE_URL = "http://localhost:8080"
python backend/tests/api_contract_test.py
k6 run infra/load/smoke.js
```
