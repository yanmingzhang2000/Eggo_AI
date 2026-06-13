from fastapi import APIRouter

from app.api.v1.endpoints import news, tasks, health, ai

api_router = APIRouter()
api_router.include_router(health.router, prefix="/health", tags=["健康检查"])
api_router.include_router(ai.router, prefix="/ai", tags=["AI 过滤"])
api_router.include_router(news.router, prefix="/news", tags=["AI 新闻"])
api_router.include_router(tasks.router, prefix="/tasks", tags=["异步任务"])
