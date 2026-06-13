from datetime import datetime

from sqlalchemy import String, SmallInteger, Text, DateTime
from sqlalchemy.dialects.postgresql import UUID, ARRAY
from sqlalchemy.orm import Mapped, mapped_column

from app.models.base import Base, UUIDMixin


class AiNews(UUIDMixin, Base):
    __tablename__ = "ai_news"

    source: Mapped[str] = mapped_column(String(64), nullable=False)
    source_url: Mapped[str | None] = mapped_column(String(1024))
    title: Mapped[str] = mapped_column(String(512), nullable=False)
    summary: Mapped[str | None] = mapped_column(Text)
    content: Mapped[str | None] = mapped_column(Text)
    related_funds: Mapped[list[str]] = mapped_column(ARRAY(UUID(as_uuid=True)), default=list)
    sentiment: Mapped[int | None] = mapped_column(SmallInteger)
    importance: Mapped[int | None] = mapped_column(SmallInteger)
    tags: Mapped[list[str]] = mapped_column(ARRAY(String(64)), default=list)
    published_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True))
    processed_at: Mapped[datetime] = mapped_column(DateTime(timezone=True))
