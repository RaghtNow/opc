# 家校协同成绩洞察平台 后端进展说明 V1

更新日期：2026-06-15

## 当前状态

后端 Go 服务骨架已创建并通过本地构建验证。

目录：
- `apps/edu-insight-api`

当前已具备：
- Go module
- Gin HTTP 服务入口
- 基础配置加载
- 健康检查接口 `/health`
- 基础信息接口 `/api/meta`
- Dockerfile
- migration 目录

## 当前目标

后续不再让前端持续积累业务逻辑，优先把 `考试与成绩` 模块迁到后端接口驱动。

## 下一步建议

### Step 1
实现考试列表和考试详情接口：
- `GET /api/exams`
- `GET /api/exams/:id`

### Step 2
实现成绩导入接口：
- `POST /api/exams/import`

### Step 3
实现成绩修改接口：
- `PATCH /api/exams/:id/scores/:scoreId`

## 当前结论

仓库已经从“只有前端原型”进入“前端 + 后端骨架并行推进”阶段。
