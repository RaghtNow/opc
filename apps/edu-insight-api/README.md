# edu-insight-api

OPC 统一后端服务骨架。

当前优先承载家校协同成绩洞察平台的业务域，但目标不是单项目后端，而是：

- 一个 Go 后端
- 多个业务域
- 多个前端项目接入
- 统一部署与统一配置管理

## 本地运行

```bash
go run ./cmd/server
```

默认监听：

```text
http://localhost:8088
```

健康检查：

```text
GET /health
```

基础信息：

```text
GET /api/meta
```

## 本地 MySQL

本地开发默认使用 MySQL：

```text
database: edu_insight_local
user: edu_insight
dsn: edu_insight:edu_insight_local_123@tcp(127.0.0.1:3306)/edu_insight_local?parseTime=true&charset=utf8mb4&loc=Local
```

启动示例：

```bash
DB_DSN='edu_insight:edu_insight_local_123@tcp(127.0.0.1:3306)/edu_insight_local?parseTime=true&charset=utf8mb4&loc=Local' go run ./cmd/server
```

如果不设置 `DB_DSN`，服务会退回内存模式，方便 CI 或无数据库环境构建验证。
