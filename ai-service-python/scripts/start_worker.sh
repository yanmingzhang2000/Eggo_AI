#!/bin/bash
# 启动 Celery Worker（处理异步任务）
celery -A app.tasks.celery_app worker --loglevel=info --concurrency=4
