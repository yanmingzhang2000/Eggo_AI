"""将爬虫原始数据格式化为 LLM 可消费的结构化 Prompt 片段。"""

import logging
from dataclasses import dataclass, field, asdict
from datetime import date, datetime

logger = logging.getLogger(__name__)


# ── 数据容器 ───────────────────────────────────────────────────
@dataclass
class FundSnapshot:
    """单只基金当日快照"""
    fund_code: str
    fund_name: str
    nav_date: str              # YYYY-MM-DD
    unit_nav: float
    acc_nav: float | None
    daily_return_pct: float    # 日涨跌幅 %
    max_drawdown_pct: float    # 近 N 日最大回撤 %
    drawdown_period_days: int  # 回撤计算窗口天数
    high_nav: float            # 窗口内最高净值
    low_nav: float             # 窗口内最低净值


@dataclass
class NewsItem:
    """单条财经快讯"""
    title: str
    source: str
    url: str
    published_at: str          # ISO 格式
    summary: str = ""


@dataclass
class DailyDigest:
    """每日汇总，喂给大模型的完整上下文"""
    digest_date: str
    funds: list[FundSnapshot] = field(default_factory=list)
    news: list[NewsItem] = field(default_factory=list)

    def to_dict(self) -> dict:
        return asdict(self)


# ── 格式化函数 ─────────────────────────────────────────────────
def format_fund_snapshot(snapshot: FundSnapshot) -> str:
    """单只基金 → 可读文本块"""
    arrow = "↑" if snapshot.daily_return_pct >= 0 else "↓"
    dd_sign = "⚠️" if snapshot.max_drawdown_pct < -5 else ""
    return (
        f"【{snapshot.fund_name}（{snapshot.fund_code}）】\n"
        f"  日期: {snapshot.nav_date}\n"
        f"  单位净值: {snapshot.unit_nav:.4f}\n"
        f"  日涨跌幅: {arrow} {snapshot.daily_return_pct:+.2f}%\n"
        f"  近{snapshot.drawdown_period_days}日最大回撤: {snapshot.max_drawdown_pct:.2f}% {dd_sign}\n"
        f"  区间最高/最低: {snapshot.high_nav:.4f} / {snapshot.low_nav:.4f}"
    )


def format_news_item(item: NewsItem) -> str:
    """单条新闻 → 可读文本行"""
    return f"- [{item.source}] {item.title}  ({item.published_at})\n  {item.url}"


def build_llm_digest(digest: DailyDigest) -> str:
    """构建完整的 LLM Prompt 片段，包含基金行情 + 财经快讯。"""
    parts: list[str] = []

    # 标题
    parts.append(f"===== 鸡生蛋 · 每日行情简报 ({digest.digest_date}) =====\n")

    # 基金部分
    if digest.funds:
        parts.append("── 基金净值与涨跌 ──")
        for snap in digest.funds:
            parts.append(format_fund_snapshot(snap))
            parts.append("")  # 空行分隔
    else:
        parts.append("（今日无基金数据）\n")

    # 新闻部分
    if digest.news:
        parts.append("── 今日财经快讯 ──")
        for item in digest.news:
            parts.append(format_news_item(item))
        parts.append("")
    else:
        parts.append("（今日无重要快讯）\n")

    # 任务指令
    parts.append(
        "── 分析任务 ──\n"
        "请基于以上行情数据与财经快讯，完成以下分析：\n"
        "1. 今日市场情绪总结（利好/利空/中性）\n"
        "2. 各基金表现点评与短期趋势判断\n"
        "3. 是否存在买卖信号？如有，请给出建议操作及理由\n"
        "4. 需要重点关注的风险提示"
    )

    text = "\n".join(parts)
    logger.info("已生成 LLM 简报，共 %d 字符", len(text))
    return text


def build_fund_snapshot(
    fund_code: str,
    fund_name: str,
    nav_date: date,
    nav_history: list[dict],
) -> FundSnapshot | None:
    """从净值历史列表构建 FundSnapshot。

    Args:
        nav_history: 按日期升序排列的 [{"nav_date": "YYYY-MM-DD", "unit_nav": 1.2345}, ...]
    """
    if not nav_history:
        logger.warning("基金 %s 无净值历史数据", fund_code)
        return None

    latest = nav_history[-1]
    unit_nav = float(latest["unit_nav"])
    acc_nav = float(latest.get("acc_nav", 0)) or None

    # 日涨跌幅
    if len(nav_history) >= 2:
        prev_nav = float(nav_history[-2]["unit_nav"])
        daily_return = (unit_nav - prev_nav) / prev_nav * 100 if prev_nav else 0.0
    else:
        daily_return = 0.0

    # 最大回撤
    navs = [float(h["unit_nav"]) for h in nav_history]
    peak = navs[0]
    max_dd = 0.0
    for n in navs:
        if n > peak:
            peak = n
        dd = (n - peak) / peak * 100
        if dd < max_dd:
            max_dd = dd

    return FundSnapshot(
        fund_code=fund_code,
        fund_name=fund_name,
        nav_date=str(nav_date),
        unit_nav=unit_nav,
        acc_nav=acc_nav,
        daily_return_pct=round(daily_return, 4),
        max_drawdown_pct=round(max_dd, 4),
        drawdown_period_days=len(nav_history),
        high_nav=max(navs),
        low_nav=min(navs),
    )
