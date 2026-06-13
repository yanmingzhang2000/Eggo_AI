"""基金净值爬虫 — 天天基金 (eastmoney.com)

功能：
  1. 抓取指定基金近 N 日净值历史
  2. 计算当日涨跌幅 & 最大回撤
  3. 格式化数据准备喂给 LLM
"""

import logging
import re
from datetime import date, datetime

from app.config import settings
from app.tasks.celery_app import app
from app.utils.http_client import http
from app.utils.llm_formatter import (
    DailyDigest,
    NewsItem,
    build_fund_snapshot,
    build_llm_digest,
)

logger = logging.getLogger(__name__)

# ── 天天基金 API 地址 ──────────────────────────────────────────
# 净值历史: https://api.fund.eastmoney.com/f10/lsjz
_FUND_NAV_API = "https://api.fund.eastmoney.com/f10/lsjz"
# 基金名称: https://fundgz.1234567.com.cn/js/{code}.js
_FUND_REALTIME_API = "https://fundgz.1234567.com.cn/js/{code}.js"

_EASTMONEY_HEADERS = {
    "Referer": "https://fundf10.eastmoney.com/",
}


def _parse_js_callback(text: str) -> dict:
    """解析天天基金 JSONP 回调格式：jsonpgz({...});"""
    match = re.search(r"jsonpgz\((.+?)\)", text)
    if match:
        import json
        return json.loads(match.group(1))
    return {}


def fetch_fund_name(fund_code: str) -> str:
    """通过实时估值接口获取基金名称。"""
    url = _FUND_REALTIME_API.format(code=fund_code)
    try:
        text = http.get_text(url, extra_headers=_EASTMONEY_HEADERS)
        data = _parse_js_callback(text)
        return data.get("name", fund_code)
    except Exception as exc:
        logger.warning("获取基金 %s 名称失败: %s", fund_code, exc)
        return fund_code


def fetch_fund_nav_history(fund_code: str, days: int = 60) -> list[dict]:
    """抓取基金近 N 日净值历史。

    Returns:
        [{"nav_date": "2025-06-12", "unit_nav": "1.2345", "acc_nav": "2.3456"}, ...]
    """
    params = {
        "fundCode": fund_code,
        "pageIndex": 1,
        "pageSize": days,
        "startDate": "",
        "endDate": "",
        "callback": "",  # 留空返回纯 JSON
    }
    data = http.get_json(
        _FUND_NAV_API,
        params=params,
        extra_headers=_EASTMONEY_HEADERS,
    )

    result = []
    for item in data.get("Data", {}).get("LSJZList", []):
        result.append({
            "nav_date": item.get("FSRQ", ""),          # 净值日期
            "unit_nav": item.get("DWJZ", "0"),         # 单位净值
            "acc_nav": item.get("LJJZ", "0"),          # 累计净值
            "daily_return": item.get("JZZZL", "0"),    # 日增长率 %
        })
    logger.info("基金 %s 获取到 %d 条净值记录", fund_code, len(result))
    return result


# ── Celery 任务 ────────────────────────────────────────────────
@app.task(
    name="crawler.fund_nav",
    bind=True,
    max_retries=2,
    default_retry_delay=60,
    acks_late=True,
)
def crawl_fund_nav(self, fund_code: str) -> dict:
    """抓取单只基金净值 + 涨跌幅 + 最大回撤，返回结构化快照。"""
    days = settings.CRAWL_NAV_DAYS
    logger.info("[fund_nav] 开始抓取基金 %s 近 %d 日净值", fund_code, days)

    try:
        fund_name = fetch_fund_name(fund_code)
        nav_history = fetch_fund_nav_history(fund_code, days)

        if not nav_history:
            logger.error("基金 %s 无净值数据", fund_code)
            return {"error": f"基金 {fund_code} 无净值数据"}

        today = date.today()
        snapshot = build_fund_snapshot(fund_code, fund_name, today, nav_history)

        if snapshot is None:
            return {"error": f"基金 {fund_code} 快照构建失败"}

        result = {
            "fund_code": snapshot.fund_code,
            "fund_name": snapshot.fund_name,
            "nav_date": snapshot.nav_date,
            "unit_nav": snapshot.unit_nav,
            "daily_return_pct": snapshot.daily_return_pct,
            "max_drawdown_pct": snapshot.max_drawdown_pct,
            "drawdown_period_days": snapshot.drawdown_period_days,
            "high_nav": snapshot.high_nav,
            "low_nav": snapshot.low_nav,
        }
        logger.info(
            "[fund_nav] %s(%s) 净值=%.4f 日涨跌=%.2f%% 最大回撤=%.2f%%",
            fund_name, fund_code, snapshot.unit_nav,
            snapshot.daily_return_pct, snapshot.max_drawdown_pct,
        )
        return result

    except Exception as exc:
        logger.exception("[fund_nav] 基金 %s 抓取失败", fund_code)
        raise self.retry(exc=exc)


@app.task(name="crawler.fund_nav_all")
def crawl_fund_nav_all() -> str:
    """定时任务入口：遍历所有目标基金，汇总后生成 LLM 简报。"""
    codes = [c.strip() for c in settings.CRAWL_FUND_CODES.split(",") if c.strip()]
    logger.info("[fund_nav_all] 开始批量抓取，共 %d 只基金: %s", len(codes), codes)

    digest = DailyDigest(digest_date=date.today().isoformat())

    for code in codes:
        try:
            snapshot_data = crawl_fund_nav(code)
            if "error" not in snapshot_data:
                from app.utils.llm_formatter import FundSnapshot
                snap = FundSnapshot(**snapshot_data)
                digest.funds.append(snap)
            else:
                logger.warning("基金 %s 抓取异常: %s", code, snapshot_data.get("error"))
        except Exception as exc:
            logger.exception("基金 %s 任务执行失败", code)

    # 生成 LLM 简报
    llm_text = build_llm_digest(digest)
    logger.info("[fund_nav_all] 批量抓取完成，生成简报 %d 字符", len(llm_text))

    # TODO: 持久化到 DB 或推送到消息队列供 AI 消费
    return llm_text
