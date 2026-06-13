from datetime import datetime
from decimal import Decimal

from sqlalchemy import String, SmallInteger, Numeric, Text, DateTime, ForeignKey, CheckConstraint
from sqlalchemy.dialects.postgresql import UUID, ARRAY
from sqlalchemy.orm import Mapped, mapped_column

from app.models.base import Base, UUIDMixin, TimestampMixin


class TradeSignal(UUIDMixin, TimestampMixin, Base):
    __tablename__ = "trade_signals"
    __table_args__ = (
        CheckConstraint("signal_type IN ('BUY', 'SELL', 'HOLD')", name="chk_signal_type"),
        CheckConstraint("confidence >= 0 AND confidence <= 100", name="chk_confidence"),
    )

    user_id: Mapped[str] = mapped_column(UUID(as_uuid=True), ForeignKey("users.id", ondelete="CASCADE"), nullable=False)
    fund_id: Mapped[str] = mapped_column(UUID(as_uuid=True), ForeignKey("funds.id", ondelete="CASCADE"), nullable=False)
    signal_type: Mapped[str] = mapped_column(String(8), nullable=False)
    strategy: Mapped[str] = mapped_column(String(64), nullable=False)
    confidence: Mapped[Decimal] = mapped_column(Numeric(5, 2), nullable=False)
    reason: Mapped[str | None] = mapped_column(Text)
    related_news_ids: Mapped[list[str]] = mapped_column(ARRAY(UUID(as_uuid=True)), default=list)
    target_amount: Mapped[Decimal | None] = mapped_column(Numeric(14, 2))
    execute_before: Mapped[datetime | None] = mapped_column(DateTime(timezone=True))
    status: Mapped[int] = mapped_column(SmallInteger, default=0)
    executed_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True))
