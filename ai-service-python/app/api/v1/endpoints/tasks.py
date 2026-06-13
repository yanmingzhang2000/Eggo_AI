from fastapi import APIRouter

router = APIRouter()


@router.get("/{task_id}")
async def get_task_status(task_id: str):
    """查询异步任务状态"""
    # TODO: 从 Celery / Redis 查询任务状态
    return {"task_id": task_id, "status": "PENDING", "result": None}
