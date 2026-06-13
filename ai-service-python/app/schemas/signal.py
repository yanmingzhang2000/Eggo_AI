from decimal import Decimal

from pydantic import BaseModel


class SignalGenerateRequest(BaseModel):
    fund_code: str
    strategy: str


class SignalItem(BaseModel):
    id: str
    fund_code: str
    signal_type: str  # BUY / SELL / HOLD
    strategy: str
    confidence: Decimal
    reason: str | None = None
    status: int
