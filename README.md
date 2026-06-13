# 鸡生蛋 Eggo · 一键部署指南

> 智能基金决策 Agent 平台

## 架构总览

```
                    ┌─────────────┐
                    │   Nginx     │
                    │  :80 / :443 │
                    └──────┬──────┘
                           │
              ┌────────────┴────────────┐
              │                         │
              ▼                         ▼
    ┌─────────────────┐       ┌─────────────────┐
    │  Go API  :8080  │       │ Python AI :8000 │
    │  核心决策引擎    │       │  AI + 爬虫服务   │
    └────────┬────────┘       └────────┬────────┘
             │                         │
             │                         │
             ▼                         ▼
    ┌─────────────────┐       ┌─────────────────┐
    │  MySQL 8.0      │       │  Redis 7        │
    │  :3306          │       │  :6379          │
    └─────────────────┘       └─────────────────┘
```

## 前置要求

- Docker 24.0+
- Docker Compose v2.20+
- 4GB+ 可用内存

## 快速启动

### 1. 克隆项目

```bash
git clone https://github.com/your-org/eggo.git
cd eggo
```

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env`，填入你的 Gemini API Key：

```bash
OPENAI_API_KEY=你的_Gemini_API_Key
```

### 3. 创建 SSL 证书目录（开发环境可用自签名证书）

```bash
mkdir -p deploy/ssl

# 生成自签名证书（仅开发环境）
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout deploy/ssl/privkey.pem \
  -out deploy/ssl/fullchain.pem \
  -subj "/CN=localhost"
```

### 4. 一键启动

```bash
docker compose up -d
```

### 5. 查看状态

```bash
# 查看所有容器状态
docker compose ps

# 查看日志
docker compose logs -f go-api      # Go API 日志
docker compose logs -f python-ai   # Python AI 日志
docker compose logs -f celery-worker  # Celery Worker 日志
```

### 6. 验证服务

```bash
# 健康检查
curl http://localhost:8080/api/v1/egg/status
curl http://localhost:8000/api/v1/health

# 通过 Nginx（HTTPS）
curl -k https://localhost/api/v1/egg/status
```

## 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| Nginx | 80, 443 | HTTP/HTTPS 入口 |
| Go API | 8080 | 核心业务 API |
| Python AI | 8000 | AI 微服务 |
| MySQL | 3306 | 主数据库 |
| Redis | 6379 | 消息队列 + 缓存 |

## 常用命令

```bash
# 启动所有服务
docker compose up -d

# 停止所有服务
docker compose down

# 停止并删除数据卷（慎用！会清空数据库）
docker compose down -v

# 重建并启动（代码更新后）
docker compose up -d --build

# 重启单个服务
docker compose restart go-api

# 进入容器调试
docker compose exec go-api sh
docker compose exec python-ai bash
docker compose exec mysql mysql -u eggo -p jishengdan
```

## 目录结构

```
eggo/
├── docker-compose.yml          # Docker 编排
├── .env                        # 环境变量（需自行创建）
├── .env.example                # 环境变量模板
├── backend-go/                 # Go 后端
│   ├── Dockerfile
│   └── ...
├── ai-service-python/          # Python AI 服务
│   ├── Dockerfile
│   └── ...
├── database/
│   ├── init/                   # MySQL 初始化脚本
│   │   └── 01_schema.sql
│   └── conf/
│       └── my.cnf              # MySQL 配置
├── deploy/
│   ├── nginx/
│   │   └── nginx.conf          # Nginx 配置
│   └── ssl/                    # SSL 证书目录
│       ├── fullchain.pem
│       └── privkey.pem
└── eggo-app/                   # uni-app 前端
```

## 生产环境注意事项

### 1. SSL 证书

生产环境请使用正式 SSL 证书（如 Let's Encrypt）：

```bash
# 安装 certbot
sudo apt install certbot

# 获取证书
sudo certbot certonly --standalone -d eggo.example.com

# 复制到项目目录
sudo cp /etc/letsencrypt/live/eggo.example.com/fullchain.pem deploy/ssl/
sudo cp /etc/letsencrypt/live/eggo.example.com/privkey.pem deploy/ssl/
```

### 2. 修改 Nginx 域名

编辑 `deploy/nginx/nginx.conf`，将 `server_name` 改为你的域名：

```nginx
server_name eggo.example.com;
```

### 3. 数据库备份

```bash
# 备份
docker compose exec mysql mysqldump -u root -p jishengdan > backup_$(date +%Y%m%d).sql

# 恢复
docker compose exec -T mysql mysql -u root -p jishengdan < backup_20250613.sql
```

### 4. 日志管理

```bash
# 查看实时日志
docker compose logs -f --tail=100

# 日志文件位置
/var/lib/docker/volumes/鸡生蛋_nginx_logs/_data/access.log
/var/lib/docker/volumes/鸡生蛋_nginx_logs/_data/error.log
```

## 故障排查

### MySQL 启动失败

```bash
# 查看 MySQL 日志
docker compose logs mysql

# 检查内存是否充足
docker stats
```

### Go API 连接数据库失败

```bash
# 检查 MySQL 是否就绪
docker compose exec mysql mysqladmin ping -u root -p

# 检查网络
docker compose exec go-api ping mysql
```

### Python AI 服务异常

```bash
# 查看详细日志
docker compose logs python-ai

# 进入容器调试
docker compose exec python-ai bash
```

## 更新日志

- v0.1.0 - 初始版本，基础架构搭建
