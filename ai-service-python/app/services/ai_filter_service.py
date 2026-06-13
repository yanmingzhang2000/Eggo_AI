"""AI 新闻过滤与摘要服务 — 基于 Gemini (OpenAI 兼容接口)

核心功能：
  1. 读取大量财经快讯，过滤掉与用户持仓/重仓行业不相关的噪音
  2. 对相关新闻用一句话提炼关联性
  3. 输出结构化 JSON（情感 / 重要性 / 标签 / 关联基金）
"""

import json
import logging
import time
from typing import Any

from openai import OpenAI, APIConnectionError, APITimeoutError, RateLimitError, APIStatusError

from app.config import settings
from app.schemas.ai import (
    AIProcessRequest,
    AIProcessResponse,
    FilteredNewsItem,
    FundInfo,
    RawNewsItem,
)

logger = logging.getLogger(__name__)

# ── System Prompt ──────────────────────────────────────────────
_SYSTEM_PROMPT = """你是一位专业的金融新闻分析师，专门为个人基金投资者提供新闻过滤服务。

## 你的核心任务
从一批财经快讯中，严格筛选出与用户持仓基金 **直接相关** 的新闻，剔除所有噪音。

## 相关性判断标准（必须严格遵守）
一条新闻 **只有满足以下任一条件** 才算「相关」：
1. 新闻直接提及用户持有的某只基金名称或代码
2. 新闻涉及该基金的重仓股票（前十大持仓）
3. 新闻涉及该基金的核心投资行业/板块
4. 新闻涉及影响该基金净值的重大宏观政策（如利率决议、行业监管政策）
5. 新闻涉及该基金的基金经理或基金公司的重大动态

## 必须过滤掉的噪音类型
- 与用户持仓完全无关的行业/个股新闻
- 纯娱乐/社会新闻
- 重复或高度相似的新闻（只保留最核心的一条）
- 过于笼统的市场综述（如"今日A股涨跌互现"且无具体关联）

## 输出要求
1. **sentiment**（情感）：-1=利空 0=中性 1=利好，必须基于对基金净值的实际影响判断
2. **importance**（重要性）：1~5 分
   - 5分：直接影响基金净值的重大事件（如重仓股暴涨/暴跌、政策突变）
   - 4分：较重要的行业/个股动态
   - 3分：有一定参考价值的信息
   - 2分：轻微相关，仅供参考
   - 1分：边缘相关
3. **relevance_reason**（关联性说明）：必须用一句话清晰说明"为什么这条新闻和你的基金有关"，必须包含具体的基金名称和关联逻辑
4. **related_funds**：关联的基金代码列表
5. **tags**：标签分类，如 ["政策","利率","行业","个股","宏观","基金经理"]

## 输入数据格式
你会收到两部分数据：
- **user_funds**：用户持仓基金列表，包含基金代码、名称、类型、重仓股票和行业
- **news_list**：待筛选的新闻列表

## 输出格式
严格按照以下 JSON Schema 输出，不要输出任何额外文字：
{
  "filtered_news": [
    {
      "original_id": "原始新闻ID",
      "title": "新闻标题",
      "source": "来源",
      "url": "链接",
      "published_at": "发布时间",
      "sentiment": -1/0/1,
      "importance": 1-5,
      "relevance_reason": "一句话说明为什么这条新闻和你的基金有关",
      "related_funds": ["基金代码"],
      "tags": ["标签"]
    }
  ],
  "market_sentiment": "整体市场情绪概要（1-2句话）"
}

如果所有新闻都不相关，返回 filtered_news 为空数组，并在 market_sentiment 中说明。
"""


def _build_user_message(request: AIProcessRequest) -> str:
    """将请求数据格式化为 LLM 可读的用户消息。"""
    parts: list[str] = []

    # 用户持仓基金信息
    parts.append("## 用户持仓基金\n")
    for i, fund in enumerate(request.user_funds, 1):
        parts.append(f"### 基金 {i}: {fund.fund_name}（{fund.fund_code}）")
        parts.append(f"- 类型: {fund.fund_type}")
        if fund.top_holdings:
            parts.append(f"- 重仓股票: {', '.join(fund.top_holdings)}")
        if fund.top_sectors:
            parts.append(f"- 重仓行业: {', '.join(fund.top_sectors)}")
        parts.append("")

    # 待筛选新闻
    parts.append("## 待筛选财经快讯\n")
    parts.append(f"共 {len(request.news_list)} 条新闻：\n")
    for i, news in enumerate(request.news_list, 1):
        parts.append(f"### 新闻 {i}")
        parts.append(f"- ID: {news.id}")
        parts.append(f"- 来源: {news.source}")
        parts.append(f"- 标题: {news.title}")
        if news.summary:
            parts.append(f"- 摘要: {news.summary[:300]}")
        if news.published_at:
            parts.append(f"- 发布时间: {news.published_at}")
        parts.append("")

    parts.append("请严格按照要求筛选并输出 JSON。")
    return "\n".join(parts)


def _parse_llm_response(raw: str) -> dict[str, Any]:
    """解析 LLM 返回的 JSON，容错处理 markdown 代码块包裹的情况。"""
    text = raw.strip()
    # 去除 markdown 代码块包裹
    if text.startswith("```"):
        # 去掉首行 ```json 和末尾 ```
        lines = text.split("\n")
        if lines[0].startswith("```"):
            lines = lines[1:]
        if lines and lines[-1].strip() == "```":
            lines = lines[:-1]
        text = "\n".join(lines).strip()

    try:
        return json.loads(text)
    except json.JSONDecodeError as exc:
        logger.error("JSON 解析失败: %s\n原始文本: %s", exc, text[:500])
        raise ValueError(f"LLM 返回的 JSON 格式无效: {exc}") from exc


def filter_and_summarize_news(
    raw_news_list: list[RawNewsItem],
    user_funds_info: list[FundInfo],
) -> AIProcessResponse:
    """核心函数：调用 Gemini 过滤新闻并提炼关联性。

    Args:
        raw_news_list: 爬虫抓取的原始新闻列表
        user_funds_info: 用户持仓基金信息

    Returns:
        AIProcessResponse: 结构化的过滤结果

    Raises:
        ValueError: LLM 返回格式异常
        RuntimeError: API 调用失败
    """
    start_time = time.monotonic()
    model = settings.LLM_MODEL

    logger.info(
        "[AI] 开始过滤新闻 — 新闻数=%d, 基金数=%d, 模型=%s",
        len(raw_news_list),
        len(user_funds_info),
        model,
    )

    # 构造请求
    request = AIProcessRequest(news_list=raw_news_list, user_funds=user_funds_info)
    user_message = _build_user_message(request)

    logger.debug("[AI] System Prompt 长度=%d 字符", len(_SYSTEM_PROMPT))
    logger.debug("[AI] User Message 长度=%d 字符", len(user_message))

    # 初始化 OpenAI 客户端（指向 Gemini 网关）
    client = OpenAI(
        api_key=settings.OPENAI_API_KEY,
        base_url=settings.OPENAI_BASE_URL,
        timeout=60.0,
    )

    # 调用 LLM
    try:
        logger.info("[AI] 正在调用 Gemini API...")
        response = client.chat.completions.create(
            model=model,
            messages=[
                {"role": "system", "content": _SYSTEM_PROMPT},
                {"role": "user", "content": user_message},
            ],
            response_format={"type": "json_object"},
            temperature=settings.LLM_TEMPERATURE,
            max_tokens=settings.LLM_MAX_TOKENS,
        )
        logger.info("[AI] Gemini API 调用成功")

    except APITimeoutError as exc:
        logger.error("[AI] Gemini API 超时: %s", exc)
        raise RuntimeError("AI 服务超时，请稍后重试") from exc
    except RateLimitError as exc:
        logger.error("[AI] Gemini API 限流: %s", exc)
        raise RuntimeError("AI 服务繁忙，请稍后重试") from exc
    except APIConnectionError as exc:
        logger.error("[AI] Gemini API 连接失败: %s", exc)
        raise RuntimeError("无法连接到 AI 服务") from exc
    except APIStatusError as exc:
        logger.error("[AI] Gemini API 返回异常状态: %d — %s", exc.status_code, exc.message)
        raise RuntimeError(f"AI 服务异常 (HTTP {exc.status_code})") from exc

    # 提取响应内容
    raw_content = response.choices[0].message.content
    if not raw_content:
        logger.error("[AI] Gemini 返回空内容")
        raise RuntimeError("AI 服务返回空结果")

    usage = response.usage
    logger.info(
        "[AI] Token 用量 — prompt=%d, completion=%d, total=%d",
        usage.prompt_tokens if usage else 0,
        usage.completion_tokens if usage else 0,
        usage.total_tokens if usage else 0,
    )

    # 解析 JSON
    parsed = _parse_llm_response(raw_content)
    raw_filtered = parsed.get("filtered_news", [])
    market_sentiment = parsed.get("market_sentiment", "")

    # 转换为响应模型
    filtered_news: list[FilteredNewsItem] = []
    for item in raw_filtered:
        try:
            filtered_news.append(FilteredNewsItem(
                original_id=item.get("original_id", ""),
                title=item.get("title", ""),
                source=item.get("source", ""),
                url=item.get("url", ""),
                published_at=item.get("published_at", ""),
                sentiment=int(item.get("sentiment", 0)),
                importance=int(item.get("importance", 3)),
                relevance_reason=item.get("relevance_reason", ""),
                related_funds=item.get("related_funds", []),
                tags=item.get("tags", []),
            ))
        except Exception as exc:
            logger.warning("[AI] 跳过无效条目: %s — %s", item, exc)

    elapsed_ms = (time.monotonic() - start_time) * 1000

    logger.info(
        "[AI] 过滤完成 — 输入=%d, 输出=%d, 耗时=%.0fms",
        len(raw_news_list),
        len(filtered_news),
        elapsed_ms,
    )

    return AIProcessResponse(
        total_input=len(raw_news_list),
        total_filtered=len(filtered_news),
        filtered_news=filtered_news,
        market_sentiment=market_sentiment,
        processing_time_ms=round(elapsed_ms, 2),
        model_used=model,
    )
