# 多阶段构建 - 依赖安装阶段
FROM node:20-alpine AS deps

WORKDIR /app

COPY package*.json ./
RUN npm install

FROM node:20-alpine AS builder

WORKDIR /app

ARG VITE_EDU_INSIGHT_API_BASE=http://127.0.0.1:8088/api
ENV VITE_EDU_INSIGHT_API_BASE=${VITE_EDU_INSIGHT_API_BASE}

COPY --from=deps /app/node_modules ./node_modules
COPY package*.json ./
COPY . .

RUN npx vite build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
