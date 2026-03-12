#!/bin/bash

# OPC 服务器部署脚本
# 用途：在云服务器上初始化和部署 OPC 应用

set -e

echo "🚀 开始部署 OPC 应用..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置变量
APP_DIR="/opt/opc"
DOCKER_USER="opc"
COMPOSE_FILE="${APP_DIR}/docker-compose.yml"
NGINX_DIR="${APP_DIR}/nginx"

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}请使用 root 权限运行此脚本${NC}"
    exit 1
fi

# 创建应用目录
echo -e "${YELLOW}📁 创建应用目录...${NC}"
mkdir -p ${APP_DIR}
mkdir -p ${NGINX_DIR}/{conf.d,ssl,logs}
mkdir -p ${APP_DIR}/ssl-certs

# 创建 Docker 网络
echo -e "${YELLOW}🌐 创建 Docker 网络...${NC}"
docker network create opc-network 2>/dev/null || echo "网络已存在"

# 复制配置文件（如果存在）
if [ -f "./docker-compose.yml" ]; then
    echo -e "${YELLOW}📋 复制 docker-compose.yml...${NC}"
    cp ./docker-compose.yml ${COMPOSE_FILE}
fi

if [ -d "./nginx" ]; then
    echo -e "${YELLOW}📋 复制 Nginx 配置...${NC}"
    cp -r ./nginx/* ${NGINX_DIR}/
fi

# 设置权限
chown -R ${DOCKER_USER}:${DOCKER_USER} ${APP_DIR}

echo -e "${GREEN}✅ 目录结构创建完成！${NC}"
echo ""
echo -e "${YELLOW}下一步：${NC}"
echo "1. 配置 SSL 证书（使用 Certbot 获取 Let's Encrypt 证书）"
echo "2. 修改 nginx.conf 中的域名"
echo "3. 运行: docker-compose up -d"
echo ""
echo -e "${YELLOW}获取 SSL 证书（示例）：${NC}"
echo "certbot certonly --nginx -d your-domain.com"
echo "cp /etc/letsencrypt/live/your-domain.com/fullchain.pem ${NGINX_DIR}/ssl/"
echo "cp /etc/letsencrypt/live/your-domain.com/privkey.pem ${NGINX_DIR}/ssl/"
