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
