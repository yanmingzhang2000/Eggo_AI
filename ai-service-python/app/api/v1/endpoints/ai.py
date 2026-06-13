"""AI 新闻过滤路由"""

import logging

from fastapi import APIRouter, HTTPException, status

from app.schemas.ai import AIProcessRequest, AIProcessResponse
from app.services.ai_filter_service import filter_and_summarize_news

logger = logging.getLogger(__name__)

router = APIRouter()


@router.post(
    "/process",
    response_model=AIProcessResponse,
    status_code=status.HTTP_200_OK,
    summary="AI 新闻过滤",
    description="接收爬虫抓取的财经快讯和用户持仓基金信息，调用 Gemini 过滤噪音并提炼关联性。",
)
async def process_news(request: AIProcessRequest) -> AIProcessResponse:
    """
    POST /api/v1/ai/process

    请求体:
    - news_list: 爬虫抓取的原始新闻列表（1~200 条）
    - user_funds: 用户持仓基金列表（含重仓股票/行业）

    返回:
    - filtered_news: 过滤后的新闻（含情感、重要性、关联性说明）
    - market_sentiment: 整体市场情绪概要
    - processing_time_ms: 处理耗时
    """
    logger.info(
        "[/ai/process] 收到请求 — 新闻数=%d, 基金数=%d",
        len(request.news_list),
        len(request.user_funds),
    )

    try:
        result = filter_and_summarize_news(
            raw_news_list=request.news_list,
            user_funds_info=request.user_funds,
        )
        logger.info(
            "[/ai/process] 处理完成 — 过滤后=%d 条, 耗时=%.0fms",
            result.total_filtered,
            result.processing_time_ms,
        )
        return result

    except ValueError as exc:
        logger.warning("[/ai/process] 参数或解析错误: %s", exc)
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail={"code": 42201, "message": f"AI 返回结果解析失败: {exc}"},
        )

    except RuntimeError as exc:
        logger.error("[/ai/process] AI 服务异常: %s", exc)
        raise HTTPException(
            status_code=status.HTTP_502_BAD_GATEWAY,
            detail={"code": 50201, "message": str(exc)},
        )

    except Exception as exc:
        logger.exception("[/ai/process] 未预期的异常")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={"code": 50001, "message": "服务器内部错误"},
        )
