# OPC 项目部署文档

## 📦 项目结构

```
workspace-developer/
├── apps/                    # 应用目录
│   └── json-parser/        # JSON 解析工具
├── infra/                   # 基础设施配置
│   └── server/             # 服务器部署配置
│       ├── deploy.sh       # 部署脚本
│       ├── update.sh       # 更新脚本
│       ├── docker-compose.yml
│       └── nginx/          # Nginx 配置
└── .github/                # GitHub Actions（需手动添加）
```

---

## 🚀 本地开发

### 前置要求
- Node.js 20+
- Docker & Docker Compose

### 启动开发环境
```bash
cd json-parser
npm install
npm run dev
```

### 使用 Docker Compose 本地运行
```bash
cd json-parser
docker-compose up -d
```

访问：http://localhost:8080

---

## ☁️ 云服务器部署

### 1. 服务器准备

购买云服务器后：
- 安装 Docker 和 Docker Compose
- 配置 SSH 访问

### 2. 初始化部署

将 `infra/server/` 目录上传到服务器并运行：

```bash
cd /opt/opc
sudo ./deploy.sh
```

### 3. 配置 SSL 证书

```bash
# 获取 Let's Encrypt 免费证书
sudo certbot certonly --nginx -d your-domain.com

# 复制证书到项目目录
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem /opt/opc/nginx/ssl/
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem /opt/opc/nginx/ssl/

# 修改 nginx.conf 中的域名
sudo nano /opt/opc/nginx/nginx.conf
```

### 4. 启动应用

```bash
cd /opt/opc
docker-compose up -d
```

### 5. 配置域名解析

在域名服务商处添加 A 记录：
```
your-domain.com → 服务器IP
```

---

## 🔄 自动化部署 (CI/CD)

### GitHub Actions 配置

由于当前 PAT 不支持 `workflow` 权限，需要重新生成包含 `workflow` 权限的 PAT：

1. GitHub → Settings → Developer settings → Personal access tokens
2. 生成新 Token，勾选权限：
   - `repo` (完整仓库访问)
   - `workflow` (工作流操作)
3. 替换 Git 凭据

然后手动添加 `.github/workflows/deploy-json-parser.yml` 到仓库。

### 配置 Secrets

在 GitHub 仓库中添加以下 Secrets：

| Secret 名称 | 说明 | 示例 |
|------------|------|------|
| `DOCKER_USERNAME` | Docker Hub 用户名 | raghtnow |
| `DOCKER_PASSWORD` | Docker Hub 密码 | xxx |
| `SERVER_HOST` | 服务器 IP | 1.2.3.4 |
| `SERVER_USER` | SSH 用户名 | root |
| `SSH_PRIVATE_KEY` | SSH 私钥 | ~/.ssh/id_rsa |

### 自动部署流程

1. 推送代码到 `dev` 分支 → 自动构建镜像
2. 推送代码到 `main` 分支 → 自动构建 + 自动部署到服务器

---

## 📝 维护命令

### 查看服务状态
```bash
cd /opt/opc
docker-compose ps
```

### 查看日志
```bash
docker-compose logs -f json-parser
docker-compose logs -f nginx
```

### 手动更新
```bash
cd /opt/opc
sudo ./update.sh
```

### 停止服务
```bash
cd /opt/opc
docker-compose down
```

---

## 🔧 故障排查

### 应用无法访问
1. 检查服务状态：`docker-compose ps`
2. 查看日志：`docker-compose logs json-parser`
3. 检查防火墙：`sudo ufw status`

### SSL 证书问题
1. 检查证书有效期：`sudo certbot certificates`
2. 续期证书：`sudo certbot renew`
3. 设置自动续期：`sudo certbot renew --dry-run`

---

## 💰 成本估算

| 项目 | 费用 |
|------|------|
| 云服务器 (2核4G) | ¥50-100/月 |
| Docker Hub | 免费 |
| GitHub Actions | 免费 |
| 域名 | ¥50-100/年 |
| SSL 证书 | 免费 (Let's Encrypt) |

**总成本**：约 ¥700-1300/年

---

## 📞 联系方式

如有问题，请联系 Developer。

---

最后更新：2026-03-12
