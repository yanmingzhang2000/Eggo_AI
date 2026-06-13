from datetime import date

from sqlalchemy import String, SmallInteger, Date
from sqlalchemy.orm import Mapped, mapped_column

from app.models.base import Base, UUIDMixin, TimestampMixin


class Fund(UUIDMixin, TimestampMixin, Base):
    __tablename__ = "funds"

    fund_code: Mapped[str] = mapped_column(String(16), unique=True, nullable=False)
    fund_name: Mapped[str] = mapped_column(String(128), nullable=False)
    fund_type: Mapped[str | None] = mapped_column(String(32))
    manager: Mapped[str | None] = mapped_column(String(64))
    custodian: Mapped[str | None] = mapped_column(String(64))
    inception_date: Mapped[date | None] = mapped_column(Date)
    benchmark: Mapped[str | None] = mapped_column(String(128))
    status: Mapped[int] = mapped_column(SmallInteger, default=1)
