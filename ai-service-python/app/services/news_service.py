from sqlalchemy.ext.asyncio import AsyncSession

from app.repositories.news_repo import NewsRepository


class NewsService:
    def __init__(self, db: AsyncSession):
        self.repo = NewsRepository(db)

    async def get_latest(self, limit: int = 20, sentiment: int | None = None):
        # TODO: 查询最新新闻
        pass

    async def analyze(self, news_id: str):
        # TODO: 调用 LLM 做情感分析 + 摘要
        pass
