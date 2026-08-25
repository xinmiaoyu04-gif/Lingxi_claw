# 后端测试

测试以 `docs/API.md` 为唯一接口契约，并复用 `data/` 中的 fixtures。

```powershell
# 仅校验提交的测试数据（无需启动服务）
python backend/tests/verify_fixtures.py

# 对运行中的后端执行 API 契约与边界测试
$env:API_BASE_URL = "http://localhost:8080"
python backend/tests/api_contract_test.py
```

`api_contract_test.py` 覆盖 Dataset 创建、任务查询、分析、计划、练习、作业提示/提交、聊天与设置接口；同时验证空参数、非法枚举、无效日期、资源不存在和统一错误信封。服务端返回的成功响应必须始终为 `{success: true, data: ..., error: null}`；错误响应必须为 `{success: false, data: null, error: {...}}`。

上传文件和真实 AI/OCR 的行为依赖实现与外部服务，保留为后端集成测试项；对应失败场景已由 fixtures 和负向 API 用例覆盖。
