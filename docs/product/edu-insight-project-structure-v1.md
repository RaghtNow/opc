# edu-insight 独立项目结构 V1

更新日期：2026-06-15

## 1. 目标

将当前教育成绩洞察产品从 `opc` 的实验性上下文中抽离，逐步收敛为独立项目：

- 名称：`edu-insight`
- 教师端：`edu-insight-teacher`
- 后端：`edu-insight-api`
- 后续小程序：`edu-insight-miniapp`

## 2. 当前仓库结构

当前 `opc` 仓库内只保留活跃代码，不再保留根目录 `edu-insight/` 空骨架，避免和 `apps/edu-insight-*` 形成重复项目。

```text
opc/
├── apps/
│   ├── edu-insight-teacher/
│   └── edu-insight-api/
└── docs/
    └── product/
```

当前约定：
- 教师端活跃代码：`apps/edu-insight-teacher`
- Go 后端活跃代码：`apps/edu-insight-api`
- 产品与架构文档：`docs/product`
- 未来小程序暂不建空目录，等进入开发时再创建真实工程

## 3. 未来独立仓库目标结构

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

## 4. 后续迁移原则

- 当前 `apps/edu-insight-teacher` 和 `apps/edu-insight-api` 仍留在 `opc` 工作区内继续开发
- 当独立仓库创建后，再整体迁移到 `edu-insight/` 结构中
- 在迁移前，命名体系和架构边界先保持一致，避免再次重命名
- 不在当前仓库保留空目录占位，只有真实进入开发的工程才创建目录
