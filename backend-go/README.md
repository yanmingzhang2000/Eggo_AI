# 鸡生蛋 - Go Backend (Gin)

## 分层架构

```
cmd/            → 程序入口
internal/       → 核心业务（不可被外部导入）
  ├─ config     → 配置加载
  ├─ middleware  → Gin 中间件
  ├─ router     → 路由注册
  ├─ controller → 请求参数绑定 & 响应
  ├─ service    → 业务逻辑
  ├─ repository → 数据库访问 (DAO)
  ├─ model      → 数据模型 (DB struct)
  ├─ dto        → 请求/响应 DTO
  └─ errors     → 统一错误码
pkg/            → 可复用工具库
api/            → API 文档 (Swagger)
scripts/        → 构建/部署脚本
```
