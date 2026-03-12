#!/bin/bash

# OPC 应用更新脚本
# 用途：拉取最新镜像并更新应用

set -e

APP_DIR="/opt/opc"

echo "🔄 开始更新 OPC 应用..."

cd ${APP_DIR}

# 拉取最新镜像
echo "📥 拉取最新 Docker 镜像..."
docker-compose pull json-parser

# 重启服务
echo "🚀 重启服务..."
docker-compose up -d json-parser

# 清理旧镜像
echo "🧹 清理旧镜像..."
docker image prune -f

echo "✅ 更新完成！"
echo ""
echo "查看服务状态："
echo "docker-compose ps"
echo ""
echo "查看日志："
echo "docker-compose logs -f json-parser"
