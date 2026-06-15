# edu-insight 独立项目结构 V1

更新日期：2026-06-15

## 1. 目标

将当前教育成绩洞察产品从 `opc` 的实验性上下文中抽离，逐步收敛为独立项目：

- 名称：`edu-insight`
- 教师端：`edu-insight-teacher`
- 后端：`edu-insight-api`
- 后续小程序：`edu-insight-miniapp`

## 2. 当前项目骨架

```text
edu-insight/
├── apps/
│   ├── teacher-web/
│   ├── miniapp/
│   └── api/
├── packages/
│   └── shared/
├── infra/
│   ├── docker/
│   └── ci/
└── docs/
```

## 3. 后续迁移原则

- 当前 `apps/edu-insight-teacher` 和 `apps/edu-insight-api` 仍留在 `opc` 工作区内继续开发
- 当独立仓库创建后，再整体迁移到 `edu-insight/` 结构中
- 在迁移前，命名体系和架构边界先保持一致，避免再次重命名
