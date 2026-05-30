# 开发规范与流程指南

版本: 1.0
更新日期: 2026-05-30

---

## 一、角色与职责

### CEO（产品决策者）
- 与业务方沟通需求、确定优先级
- 验收功能、做产品决策
- 排期、拍板技术方案

### Developer（开发工程师）
- 业务代码实现（前端 Vue、工具逻辑）
- 本地开发、自测
- 单元测试、代码质量

### DevOps（运维工程师）
- CI/CD 流水线（GitHub Actions）
- Dockerfile、docker-compose、部署脚本
- 服务器配置、nginx、SSL、Secrets

### 红线
- Developer 不碰 GitHub Actions、Dockerfile、部署脚本
- DevOps 不碰业务代码（src/ 下的 Vue/TS 文件）
- 跨领域修改必须提前沟通

---

## 二、分支策略

```
main     ← 生产分支，只接受 PR 合并
  ↑
dev      ← 开发分支，功能集成测试
  ↑
feature/xxx  ← 功能分支，从这里开发
```

### 分支命名
| 类型 | 格式 | 示例 |
|------|------|------|
| 新功能 | `feature/功能描述` | `feature/json-export` |
| 修复 | `fix/问题描述` | `fix/editor-crash` |
| 重构 | `refactor/范围` | `refactor/parser-utils` |
| 文档 | `docs/内容` | `docs/api-guide` |
| CI/CD | `ci/变更` | `ci/add-linter` |

---

## 三、开发流程

### 第1步：切分支
```bash
git checkout main && git pull origin main
git checkout -b feature/你的功能名
```

### 第2步：本地开发
```bash
# 进入对应应用
cd apps/json-editor-tool

# 安装依赖（首次）
npm install

# 启动开发服务器
npm run dev

# 构建测试（确保生产构建无报错）
npm run build
```

### 第3步：提交代码
```bash
git add .
git commit -m "feat: 添加 JSON 导出功能

- 支持导出为 CSV、YAML 格式
- 添加导出按钮到工具栏"
```

### 提交信息规范（Conventional Commits）
| 类型 | 用途 |
|------|------|
| `feat:` | 新功能 |
| `fix:` | 修复 bug |
| `refactor:` | 重构（不改行为） |
| `docs:` | 文档更新 |
| `test:` | 测试代码 |
| `ci:` | CI/CD 配置 |
| `chore:` | 杂项（依赖更新等） |

### 第4步：推送到远程
```bash
git push -u origin feature/你的功能名
```

### 第5步：创建 Pull Request
- PR 标题：`feat: 添加 JSON 导出功能`
- 描述：改了什么、为什么改、测试方式
- Reviewer：CEO
- 合并到 `main`

### 第6步：合并后自动部署
- PR 合并到 `main` → GitHub Actions 自动构建镜像
- 构建成功 → 自动 SSH 部署到服务器
- 部署完成 → 服务自动更新

---

## 四、自动化部署链路

```
git push main
    ↓
GitHub Actions: docker-xxx.yml
    ↓
构建镜像 → 推送到阿里云 ACR
    ↓
GitHub Actions: deploy-xxx.yml
    ↓
SSH 连服务器 → docker compose pull → up -d
    ↓
服务更新完成
```

### 各应用流水线
| 应用 | 构建工作流 | 部署工作流 |
|------|-----------|-----------|
| json-parser | `docker-json-parser.yml` | `deploy-json-parser.yml` |
| json-editor-tool | `docker-json-editor.yml` | `deploy-json-editor.yml` |

---

## 五、本地目录结构

```
~/opc-dev/
├── apps/
│   ├── json-editor-tool/     ← Developer 工作区
│   └── json-parser/          ← Developer 工作区
├── infra/
│   ├── scripts/              ← DevOps 工作区
│   └── server/               ← DevOps 工作区
├── .github/workflows/        ← DevOps 工作区
├── docs/                     ← 共同维护
└── memory/                   ← CEO 记忆
```

---

## 六、回滚机制

### 快速回滚（发现线上问题）
```bash
# 在服务器上执行
cd /opt/opc
docker compose pull  # 拉取上一个版本的镜像标签
docker compose up -d
docker image prune -f
```

### 通过 GitHub Actions 回滚
- 在 GitHub 仓库 → Actions → 找到之前的成功运行
- 点击 "Re-run workflow"

---

## 七、新应用接入标准

新增应用时按以下清单执行：

1. 在 `apps/` 下创建应用目录
2. 编写 `Dockerfile`（多阶段构建，nginx 托管）
3. 编写 `nginx.conf`
4. 在 `.github/workflows/` 创建：
   - `docker-<应用名>.yml`（构建推送镜像）
   - `deploy-<应用名>.yml`（SSH 部署）
5. 更新 `infra/server/docker-compose.yml` 添加服务
6. 更新本文档

---

## 八、沟通规则

- 需求变更：CEO 直接下达，Developer 评估技术可行性后执行
- 技术方案有争议：CEO 拍板
- 线上故障：DevOps 第一时间处理，同步 CEO
- 代码审查：PR 必须经过 review 才能合并

---

最后更新：2026-05-30
