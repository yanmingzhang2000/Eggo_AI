from sqlalchemy.ext.asyncio import AsyncSession

from app.models.ai_news import AiNews


class NewsRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def find_by_id(self, news_id: str) -> AiNews | None:
        # TODO
        pass

    async def find_latest(self, limit: int = 20, sentiment: int | None = None) -> list[AiNews]:
        # TODO
        pass

    async def update_analysis(self, news_id: str, summary: str, sentiment: int, tags: list[str]):
        # TODO
        pass
