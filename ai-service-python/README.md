# 鸡生蛋 - AI Service (FastAPI)

## 分层架构

```
app/
  ├─ main.py           → FastAPI 实例 & 启动
  ├─ config.py         → 配置加载 (pydantic-settings)
  ├─ deps.py           → 公共依赖注入
  ├─ api/              → 路由层 (Router)
  │   └─ v1/
  ├─ schemas/          → Pydantic 请求/响应模型 (DTO)
  ├─ services/         → 业务逻辑层
  ├─ repositories/     → 数据访问层 (DAO)
  ├─ models/           → SQLAlchemy ORM 模型
  ├─ core/             → 核心组件 (安全/日志/异常)
  ├─ tasks/            → Celery 异步任务
  │   ├─ crawler/      → 爬虫任务
  │   └─ ai/           → AI 推理任务
  └─ utils/            → 工具函数
alembic/               → 数据库迁移
tests/                 → 测试
scripts/               → 运维脚本
```
