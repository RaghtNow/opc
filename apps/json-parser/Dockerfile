# 多阶段构建 - 依赖安装阶段
FROM node:20-alpine AS deps

WORKDIR /app

# 复制 package.json 和 package-lock.json 以利用缓存层
COPY package*.json ./
RUN npm ci

# 多阶段构建 - 构建阶段
FROM node:20-alpine AS builder

WORKDIR /app

# 复制依赖
COPY --from=deps /app/node_modules ./node_modules
COPY package*.json ./

# 复制源码
COPY . .

# 构建（跳过类型检查以加快速度）
RUN npx vite build

# 生产阶段 - 使用轻量级 nginx
FROM nginx:alpine

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
