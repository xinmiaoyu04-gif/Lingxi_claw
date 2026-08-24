# 压测

安装 [k6](https://grafana.com/docs/k6/latest/set-up/install-k6/) 后运行：

```powershell
$env:API_BASE_URL = "http://localhost:8080"
k6 run infra/load/smoke.js
```

默认进行 5 个并发用户、30 秒的聊天接口冒烟压测。通过标准：失败率低于 1%，P95 响应时间低于 1 秒。请只在本地、预发布或已获授权的测试环境运行。
