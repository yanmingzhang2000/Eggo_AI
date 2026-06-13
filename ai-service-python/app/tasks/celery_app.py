from celery import Celery
from celery.schedules import crontab

from app.config import settings

app = Celery(
    "jishengdan",
    broker=settings.CELERY_BROKER_URL,
    backend=settings.CELERY_RESULT_BACKEND,
)

app.conf.update(
    task_serializer="json",
    accept_content=["json"],
    result_serializer="json",
    timezone="Asia/Shanghai",
    enable_utc=False,
    task_track_started=True,
    task_acks_late=True,
    worker_prefetch_multiplier=1,
    # 自动发现 tasks 包下的所有任务
    include=[
        "app.tasks.crawler.fund_nav",
        "app.tasks.crawler.news",
    ],
)

# ── Beat 定时调度 ──────────────────────────────────────────────
app.conf.beat_schedule = {
    # 每天 15:30 抓取基金净值 + 涨跌幅 + 最大回撤
    "crawl-fund-nav-daily": {
        "task": "crawler.fund_nav_all",
        "schedule": crontab(hour=15, minute=30),
        "args": (),
    },
    # 每天 08:00 / 12:00 / 18:00 抓取财经快讯
    "crawl-finance-news": {
        "task": "crawler.finance_news",
        "schedule": crontab(hour="8,12,18", minute=0),
        "args": (),
    },
}
