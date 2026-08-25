# 测试与 Mock 数据

本目录中的 JSON 数据遵循 `docs/DATA_SCHEMA.md` 与 `docs/API.md`，可供前端 Mock 模式、后端开发和 API 契约测试共用。

- `fixtures/`：领域对象与请求样例。
- `mock/`：可直接作为接口成功响应返回的固定响应。
- `invalid/`：用于验证输入校验与错误处理的请求样例。

不在此目录保存真实用户上传资料或密钥。运行 `python backend/tests/verify_fixtures.py` 可验证 JSON 语法、统一响应信封和关键字段。
