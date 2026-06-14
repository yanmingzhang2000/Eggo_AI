import os
from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.config import settings
from app.api.v1.router import api_router


@asynccontextmanager
async def lifespan(app: FastAPI):
    # startup: 初始化 DB 连接池、Redis、Celery 等
    yield
    # shutdown: 清理资源


def create_app() -> FastAPI:
    app = FastAPI(
        title="鸡生蛋 AI Service",
        version="0.1.0",
        docs_url="/docs",
        redoc_url="/redoc",
        lifespan=lifespan,
    )
    app.include_router(api_router, prefix="/api/v1")
    return app


app = create_app()

if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("app.main:app", host="0.0.0.0", port=port, reload=settings.DEBUG)
