# Eggo 国内金融 API 代理 — 部署指南

## 1. 前置条件

- Cloudflare 账号（免费即可）
- Node.js 18+
- Wrangler CLI

```bash
npm install -g wrangler
wrangler login
```

## 2. 部署 Worker

```bash
cd cloudflare-proxy

# 创建 KV 命名空间（可选，用于缓存）
wrangler kv:namespace create CACHE

# 将输出的 ID 填入 wrangler.toml
# [kv_namespaces]
# binding = "CACHE"
# id = "xxxxxxxxxxxxxxxxxxxxxxxx"

# 本地测试
wrangler dev

# 部署到 Cloudflare
wrangler deploy
```

部署后会得到一个 URL，类似：
```
https://eggo-fund-proxy.your-username.workers.dev
```

## 3. 配置后端环境变量

在 Render 的环境变量中添加：

```
CF_PROXY_URL=https://eggo-fund-proxy.your-username.workers.dev
CF_PROXY_KEY=eggo-proxy-2026
```

## 4. 测试

### 测试代理是否正常
```bash
# 直接测试 Worker
curl "https://your-worker.workers.dev/proxy?url=https://fundgz.1234567.com.cn/js/110011.js&key=eggo-proxy-2026"

# 测试批量请求
curl -X POST "https://your-worker.workers.dev/batch?key=eggo-proxy-2026" \
  -H "Content-Type: application/json" \
  -d '{"urls":["https://fundgz.1234567.com.cn/js/110011.js","https://fundgz.1234567.com.cn/js/161725.js"]}'

# 健康检查
curl "https://your-worker.workers.dev/health"
```

### 测试后端是否走代理
```bash
# 查看后端日志，确认请求通过代理
curl "https://your-render-app.onrender.com/api/v1/market/fund-distribution"
```

## 5. 数据流

```
用户浏览器
    ↓
Render（境外）← 后端 Go 服务
    ↓
Cloudflare Worker（全球边缘节点）
    ↓
国内金融 API（天天基金/东方财富）
    ↓
数据原路返回
```

## 6. 免费额度

| 资源 | 免费额度 | 说明 |
|------|---------|------|
| Worker 请求 | 100,000 次/天 | 足够 3 个指数 × 5 分钟刷新 × 24 小时 |
| KV 读取 | 100,000 次/天 | 缓存命中不消耗 |
| KV 写入 | 1,000 次/天 | 每个缓存键写一次 |
| KV 存储 | 1 GB | 足够 |

## 7. 自定义配置

修改 `wrangler.toml` 中的环境变量：

```toml
[vars]
API_KEY = "your-secret-key"        # 修改为你的密钥
ALLOWED_ORIGINS = "https://your-domain.com"  # 修改为你的域名
```

## 8. 注意事项

- Worker 在亚洲有节点，延迟约 50-100ms
- KV 缓存可显著降低延迟（命中时 ~0ms）
- 如果不需要缓存，可以删除 KV 配置，Worker 仍然可以工作
- 生产环境建议修改 `API_KEY`
