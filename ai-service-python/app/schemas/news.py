from datetime import datetime

from pydantic import BaseModel


class NewsItem(BaseModel):
    id: str
    source: str
    title: str
    summary: str | None = None
    sentiment: int | None = None  # -1 / 0 / 1
    importance: int | None = None  # 1~5
    tags: list[str] = []
    published_at: datetime | None = None


class NewsAnalyzeRequest(BaseModel):
    news_id: str


class NewsAnalyzeResponse(BaseModel):
    task_id: str
    status: str
