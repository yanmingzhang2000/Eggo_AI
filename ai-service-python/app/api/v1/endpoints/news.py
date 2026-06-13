from fastapi import APIRouter, Depends, Query
from sqlalchemy.ext.asyncio import AsyncSession

from app.deps import get_db

router = APIRouter()


@router.get("/latest")
async def get_latest_news(
    limit: int = Query(default=20, le=100),
    sentiment: int | None = None,
    db: AsyncSession = Depends(get_db),
):
    """获取最新 AI 过滤新闻"""
    # TODO: 调用 NewsService
    return {"items": [], "total": 0}


@router.post("/analyze")
async def analyze_news(
    news_id: str,
    db: AsyncSession = Depends(get_db),
):
    """触发单条新闻 AI 分析（情感 + 摘要）"""
    # TODO: 派发 Celery 任务
    return {"task_id": "pending", "status": "queued"}
