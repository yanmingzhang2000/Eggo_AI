from sqlalchemy.ext.asyncio import AsyncSession

from app.models.trade_signal import TradeSignal


class SignalRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def create(self, signal: TradeSignal) -> TradeSignal:
        # TODO
        pass

    async def find_by_user(self, user_id: str, status: int | None = None) -> list[TradeSignal]:
        # TODO
        pass
