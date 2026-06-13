from datetime import date
from decimal import Decimal

from sqlalchemy import Date, Numeric, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from app.models.base import Base, UUIDMixin


class FundNavDaily(UUIDMixin, Base):
    __tablename__ = "fund_nav_daily"

    fund_id: Mapped[str] = mapped_column(UUID(as_uuid=True), ForeignKey("funds.id", ondelete="CASCADE"), nullable=False)
    nav_date: Mapped[date] = mapped_column(Date, nullable=False)
    unit_nav: Mapped[Decimal] = mapped_column(Numeric(12, 4), nullable=False)
    acc_nav: Mapped[Decimal | None] = mapped_column(Numeric(12, 4))
    daily_return: Mapped[Decimal | None] = mapped_column(Numeric(8, 4))
    total_return: Mapped[Decimal | None] = mapped_column(Numeric(8, 4))
