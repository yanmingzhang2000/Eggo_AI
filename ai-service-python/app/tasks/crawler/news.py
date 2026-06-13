"""财经快讯爬虫 — 财联社电报 + RSS 源

数据源：
  1. 财联社电报 (cls.cn) — 实时快讯 API
  2. 新浪财经 RSS — finance.sina.com.cn
  3. 36氪 RSS — 36kr.com

输出：结构化 NewsItem 列表，格式化后喂给 LLM。
"""

import logging
import re
import xml.etree.ElementTree as ET
from datetime import datetime, timezone, timedelta

from app.config import settings
from app.tasks.celery_app import app
from app.utils.http_client import http
from app.utils.llm_formatter import DailyDigest, NewsItem, build_llm_digest

logger = logging.getLogger(__name__)

# 北京时间时区
_CST = timezone(timedelta(hours=8))

# ── 财联社电报 ─────────────────────────────────────────────────
_CLS_TELEGRAPH_URL = "https://www.cls.cn/nodeapi/updateTelegraphList"
_CLS_HEADERS = {
    "Referer": "https://www.cls.cn/telegraph",
    "Origin": "https://www.cls.cn",
}


def fetch_cls_telegraph(limit: int = 30) -> list[NewsItem]:
    """抓取财联社电报快讯。"""
    params = {
        "app": "CailianpressWeb",
        "os": "web",
        "sv": "8.4.6",
        "rn": str(limit),
    }
    logger.info("[cls] 开始抓取财联社电报，数量=%d", limit)

    try:
        data = http.get_json(_CLS_TELEGRAPH_URL, params=params, extra_headers=_CLS_HEADERS)
    except Exception as exc:
        logger.exception("[cls] 请求失败")
        return []

    items: list[NewsItem] = []
    for row in data.get("data", {}).get("roll_data", []):
        title = row.get("title", "").strip()
        content = row.get("content", "").strip()
        # 去除 HTML 标签
        content = re.sub(r"<[^>]+>", "", content)
        ts = row.get("ctime", 0)
        pub_time = datetime.fromtimestamp(ts, tz=_CST).isoformat() if ts else ""

        if not title and content:
            title = content[:60] + ("..." if len(content) > 60 else "")

        items.append(NewsItem(
            title=title,
            source="财联社",
            url=f"https://www.cls.cn/detail/{row.get('id', '')}",
            published_at=pub_time,
            summary=content[:200] if content else "",
        ))

    logger.info("[cls] 抓取到 %d 条电报", len(items))
    return items


# ── RSS 通用解析 ───────────────────────────────────────────────
_RSS_SOURCES = {
    "新浪财经": "https://finance.sina.com.cn/roll/index.d.html?cid=56588&page=1",
    # 以下为备用 RSS 源（部分可能需要翻墙）
    # "36氪": "https://36kr.com/feed",
    # "华尔街见闻": "https://wallstreetcn.com/rss/news",
}

# 新浪财经滚动新闻 API
_SINA_ROLL_API = "https://feed.mix.sina.com.cn/api/roll/get"
_SINA_HEADERS = {
    "Referer": "https://finance.sina.com.cn/",
}


def fetch_sina_finance_news(limit: int = 30) -> list[NewsItem]:
    """抓取新浪财经滚动新闻。"""
    params = {
        "pageid": "153",
        "lid": "2516",       # 财经新闻
        "k": "",
        "num": str(limit),
        "page": "1",
    }
    logger.info("[sina] 开始抓取新浪财经新闻，数量=%d", limit)

    try:
        data = http.get_json(_SINA_ROLL_API, params=params, extra_headers=_SINA_HEADERS)
    except Exception as exc:
        logger.exception("[sina] 请求失败")
        return []

    items: list[NewsItem] = []
    for row in data.get("result", {}).get("data", []):
        title = row.get("title", "").strip()
        url = row.get("url", "")
        summary = row.get("summary", "").strip()
        ts = int(row.get("ctime", 0))
        pub_time = datetime.fromtimestamp(ts, tz=_CST).isoformat() if ts else ""

        items.append(NewsItem(
            title=title,
            source="新浪财经",
            url=url,
            published_at=pub_time,
            summary=summary[:200],
        ))

    logger.info("[sina] 抓取到 %d 条新闻", len(items))
    return items


def fetch_rss_feed(name: str, url: str, limit: int = 20) -> list[NewsItem]:
    """通用 RSS/Atom 解析器。"""
    logger.info("[rss] 解析 %s: %s", name, url)
    try:
        text = http.get_text(url)
    except Exception as exc:
        logger.warning("[rss] %s 请求失败: %s", name, exc)
        return []

    items: list[NewsItem] = []
    try:
        root = ET.fromstring(text)
        # 兼容 RSS 2.0 和 Atom
        ns = {"atom": "http://www.w3.org/2005/Atom"}
        entries = root.findall(".//item") or root.findall(".//atom:entry", ns)

        for entry in entries[:limit]:
            title_el = entry.find("title") or entry.find("atom:title", ns)
            link_el = entry.find("link") or entry.find("atom:link", ns)
            desc_el = entry.find("description") or entry.find("atom:summary", ns)
            pub_el = (
                entry.find("pubDate")
                or entry.find("dc:date", {"dc": "http://purl.org/dc/elements/1.1/"})
                or entry.find("atom:updated", ns)
            )

            title = title_el.text.strip() if title_el is not None and title_el.text else ""
            link = link_el.text.strip() if link_el is not None and link_el.text else ""
            if not link and link_el is not None:
                link = link_el.get("href", "")
            desc = desc_el.text.strip() if desc_el is not None and desc_el.text else ""
            # 去除 HTML
            desc = re.sub(r"<[^>]+>", "", desc)[:200]
            pub_at = pub_el.text.strip() if pub_el is not None and pub_el.text else ""

            items.append(NewsItem(
                title=title,
                source=name,
                url=link,
                published_at=pub_at,
                summary=desc,
            ))
    except ET.ParseError as exc:
        logger.warning("[rss] %s XML 解析失败: %s", name, exc)

    logger.info("[rss] %s 解析到 %d 条", name, len(items))
    return items


# ── Celery 任务 ────────────────────────────────────────────────
@app.task(name="crawler.finance_news")
def crawl_finance_news() -> str:
    """定时任务入口：抓取多源财经快讯，汇总后生成 LLM 简报。"""
    logger.info("[news] 开始抓取财经快讯")

    all_news: list[NewsItem] = []

    # 1. 财联社电报
    all_news.extend(fetch_cls_telegraph(limit=30))

    # 2. 新浪财经
    all_news.extend(fetch_sina_finance_news(limit=20))

    # 3. RSS 源
    for name, url in _RSS_SOURCES.items():
        all_news.extend(fetch_rss_feed(name, url, limit=15))

    # 按时间倒序去重（标题去重）
    seen_titles: set[str] = set()
    unique_news: list[NewsItem] = []
    for item in all_news:
        key = item.title[:30]  # 前 30 字作为去重 key
        if key and key not in seen_titles:
            seen_titles.add(key)
            unique_news.append(item)

    unique_news.sort(key=lambda x: x.published_at, reverse=True)
    logger.info("[news] 去重后共 %d 条快讯", len(unique_news))

    # 构建 LLM 简报
    digest = DailyDigest(
        digest_date=datetime.now(_CST).date().isoformat(),
        news=unique_news[:50],  # 最多保留 50 条
    )
    llm_text = build_llm_digest(digest)
    logger.info("[news] 生成新闻简报 %d 字符", len(llm_text))

    # TODO: 持久化到 DB / 推送到消息队列
    return llm_text
