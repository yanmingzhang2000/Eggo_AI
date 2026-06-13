from sqlalchemy.ext.asyncio import AsyncSession

from app.repositories.signal_repo import SignalRepository


class SignalService:
    def __init__(self, db: AsyncSession):
        self.repo = SignalRepository(db)

    async def generate(self, fund_code: str, strategy: str):
        # TODO: 基于 AI + 策略生成买卖信号
        pass
