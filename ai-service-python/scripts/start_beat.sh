#!/bin/bash
# 启动 Celery Beat（定时调度器）
celery -A app.tasks.celery_app beat --loglevel=info
