from sqlalchemy import String, Integer, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from app.models.base import Base, UUIDMixin


class Watchlist(UUIDMixin, Base):
    __tablename__ = "watchlist"

    user_id: Mapped[str] = mapped_column(UUID(as_uuid=True), ForeignKey("users.id", ondelete="CASCADE"), nullable=False)
    fund_id: Mapped[str] = mapped_column(UUID(as_uuid=True), ForeignKey("funds.id", ondelete="CASCADE"), nullable=False)
    remark: Mapped[str | None] = mapped_column(String(256))
    sort_order: Mapped[int] = mapped_column(Integer, default=0)
