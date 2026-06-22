"""财经快讯爬虫 — 财联社电报 + 新浪财经 + RSS 源"""

import logging
import re
import xml.etree.ElementTree as ET
from datetime import datetime, timezone, timedelta

import httpx

from app.utils.llm_formatter import DailyDigest, NewsItem, build_llm_digest

logger = logging.getLogger(__name__)

_CST = timezone(timedelta(hours=8))

_CLS_TELEGRAPH_URL = "https://www.cls.cn/nodeapi/updateTelegraphList"
_CLS_HEADERS = {"Referer": "https://www.cls.cn/telegraph", "Origin": "https://www.cls.cn"}

_SINA_ROLL_API = "https://feed.mix.sina.com.cn/api/roll/get"
_SINA_HEADERS = {"Referer": "https://finance.sina.com.cn/"}

_RSS_SOURCES = {
    "新浪财经": "https://finance.sina.com.cn/roll/index.d.html?cid=56588&page=1",
}


def fetch_cls_telegraph(limit: int = 30) -> list[NewsItem]:
    params = {"app": "CailianpressWeb", "os": "web", "sv": "8.4.6", "rn": str(limit)}
    try:
        resp = httpx.get(_CLS_TELEGRAPH_URL, params=params, headers=_CLS_HEADERS, timeout=15)
        resp.raise_for_status()
        data = resp.json()
    except Exception:
        logger.exception("[cls] 请求失败")
        return []

    items = []
    for row in data.get("data", {}).get("roll_data", []):
        title = row.get("title", "").strip()
        content = re.sub(r"<[^>]+>", "", row.get("content", "").strip())
        ts = row.get("ctime", 0)
        pub_time = datetime.fromtimestamp(ts, tz=_CST).isoformat() if ts else ""
        if not title and content:
            title = content[:60] + ("..." if len(content) > 60 else "")
        items.append(NewsItem(
            title=title, source="财联社",
            url=f"https://www.cls.cn/detail/{row.get('id', '')}",
            published_at=pub_time, summary=content[:200],
        ))
    return items


def fetch_sina_finance_news(limit: int = 30) -> list[NewsItem]:
    params = {"pageid": "153", "lid": "2516", "k": "", "num": str(limit), "page": "1"}
    try:
        resp = httpx.get(_SINA_ROLL_API, params=params, headers=_SINA_HEADERS, timeout=15)
        resp.raise_for_status()
        data = resp.json()
    except Exception:
        logger.exception("[sina] 请求失败")
        return []

    items = []
    for row in data.get("result", {}).get("data", []):
        ts = int(row.get("ctime", 0))
        pub_time = datetime.fromtimestamp(ts, tz=_CST).isoformat() if ts else ""
        items.append(NewsItem(
            title=row.get("title", "").strip(), source="新浪财经",
            url=row.get("url", ""), published_at=pub_time,
            summary=row.get("summary", "").strip()[:200],
        ))
    return items


def fetch_rss_feed(name: str, url: str, limit: int = 20) -> list[NewsItem]:
    try:
        resp = httpx.get(url, timeout=15)
        resp.raise_for_status()
        text = resp.text
    except Exception:
        logger.warning("[rss] %s 请求失败", name)
        return []

    items = []
    try:
        root = ET.fromstring(text)
        ns = {"atom": "http://www.w3.org/2005/Atom"}
        entries = root.findall(".//item") or root.findall(".//atom:entry", ns)
        for entry in entries[:limit]:
            title_el = entry.find("title") or entry.find("atom:title", ns)
            link_el = entry.find("link") or entry.find("atom:link", ns)
            desc_el = entry.find("description") or entry.find("atom:summary", ns)
            pub_el = entry.find("pubDate") or entry.find("atom:updated", ns)
            title = title_el.text.strip() if title_el is not None and title_el.text else ""
            link = link_el.text.strip() if link_el is not None and link_el.text else ""
            if not link and link_el is not None:
                link = link_el.get("href", "")
            desc = re.sub(r"<[^>]+>", "", desc_el.text.strip() if desc_el is not None and desc_el.text else "")[:200]
            pub_at = pub_el.text.strip() if pub_el is not None and pub_el.text else ""
            items.append(NewsItem(title=title, source=name, url=link, published_at=pub_at, summary=desc))
    except ET.ParseError as exc:
        logger.warning("[rss] %s XML 解析失败: %s", name, exc)
    return items


def crawl_finance_news() -> list[NewsItem]:
    all_news = []
    all_news.extend(fetch_cls_telegraph(limit=30))
    all_news.extend(fetch_sina_finance_news(limit=20))
    for name, url in _RSS_SOURCES.items():
        all_news.extend(fetch_rss_feed(name, url, limit=15))

    seen = set()
    unique = []
    for item in all_news:
        key = item.title[:30]
        if key and key not in seen:
            seen.add(key)
            unique.append(item)
    unique.sort(key=lambda x: x.published_at, reverse=True)
    logger.info("[news] 抓取 %d 条, 去重后 %d 条", len(all_news), len(unique))
    return unique[:50]
