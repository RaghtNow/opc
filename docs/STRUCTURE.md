# 项目目录结构

```
workspace-developer/
├── apps/                    # 应用目录
│   └── json-parser/        # JSON 解析工具
│       ├── src/
│       ├── public/
│       ├── Dockerfile
│       ├── nginx.conf
│       ├── package.json
│       └── docker-compose.yml
├── infra/                   # 基础设施
│   ├── scripts/            # 部署脚本
│   │   ├── deploy.sh
│   │   └── update.sh
│   └── server/             # 服务器配置
│       ├── docker-compose.yml
│       └── nginx/
├── shared/                  # 共享代码（预留）
├── scripts/                 # 构建脚本（预留）
├── docs/                    # 项目文档
│   └── DEPLOYMENT.md
├── memory/                  # OpenClaw 状态文件
├── .github/                 # GitHub 配置
│   └── workflows/
│       └── deploy-json-parser.yml
├── AGENTS.md                # Agent 规范
├── SOUL.md                  # 项目灵魂
└── README.md                # 项目说明
```

## 目录说明

### apps/
存放所有应用代码，每个应用独立目录。

### infra/
基础设施配置，包括部署脚本和服务器配置。

### shared/
共享代码、工具类和类型定义。

### scripts/
构建、测试和发布脚本。

### docs/
项目文档，包括部署文档、API 文档等。
