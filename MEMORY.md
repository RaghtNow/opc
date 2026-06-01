# 项目长期记忆

## 代码推送规范（必须执行）

### 分支策略
- `main`：生产分支，**只接受 PR 合并**，禁止直接推送
- `dev`：开发分支，功能集成测试
- `feature/xxx`、`fix/xxx`：功能/修复分支，从这里开发

### 提交前检查清单
1. 当前是否在功能分支？（禁止在 main 上直接开发）
2. `npm run build` 是否通过？
3. Commit 是否符合 Conventional Commits？
   - `feat:`、`fix:`、`refactor:`、`docs:`、`test:`、`ci:`、`chore:`
4. Commit message 是否描述清楚改了什么、为什么改

### 推送流程
```
功能开发 → 本地构建通过 → commit → push 到功能分支 → 创建 PR → Review 通过 → 合并到 main → Actions 自动构建部署
```

### 红线
- Developer 不碰 CI/CD、Dockerfile、部署脚本
- DevOps 不碰业务代码
- 跨领域修改必须提前沟通

## 部署链路记忆
- 镜像仓库：阿里云 ACR（个人版）
- 服务器路径：`/opt/opc`
- 部署方式：GitHub Actions SSH 远程执行 `docker compose pull && up -d`
- 各应用流水线见 `docs/DEVELOPMENT_GUIDE.md`

## 项目结构
- `apps/`：业务代码（Developer 负责）
- `infra/`：部署脚本、服务器配置（DevOps 负责）
- `.github/workflows/`：CI/CD（DevOps 负责）
- `docs/`：文档（共同维护）
- `memory/`：每日工作记录

## 会话启动必读
每次新会话启动时：
1. 读取 `SOUL.md`、`USER.md`
2. 读取本文件
3. 读取 `memory/YYYY-MM-DD.md`（今天 + 昨天）
