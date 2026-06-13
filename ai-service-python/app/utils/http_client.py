import logging
import random
import time
from typing import Any

import httpx

logger = logging.getLogger(__name__)

# ── UA 池 ─────────────────────────────────────────────────────
_USER_AGENTS = [
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
]

_ACCEPT_LANGS = [
    "zh-CN,zh;q=0.9,en;q=0.8",
    "zh-CN,zh;q=0.9",
    "zh-CN,zh;q=0.8,en-US;q=0.5,en;q=0.3",
    "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
]


def _random_headers(extra: dict[str, str] | None = None) -> dict[str, str]:
    """生成随机化请求头，模拟真实浏览器指纹。"""
    headers = {
        "User-Agent": random.choice(_USER_AGENTS),
        "Accept": "application/json, text/plain, */*",
        "Accept-Language": random.choice(_ACCEPT_LANGS),
        "Accept-Encoding": "gzip, deflate, br",
        "Connection": "keep-alive",
        "Cache-Control": "no-cache",
        "Pragma": "no-cache",
    }
    if extra:
        headers.update(extra)
    return headers


def _jitter_sleep(base: float = 1.0, jitter: float = 0.5) -> None:
    """随机等待 base ± jitter 秒，模拟人类行为。"""
    delay = base + random.uniform(-jitter, jitter)
    time.sleep(max(0.1, delay))


class HttpClient:
    """封装 httpx.Client，内置反爬策略：随机 Header、自动重试、请求间隔。"""

    def __init__(
        self,
        timeout: float = 15.0,
        max_retries: int = 3,
        base_delay: float = 1.0,
    ):
        self._timeout = timeout
        self._max_retries = max_retries
        self._base_delay = base_delay

    def get_json(
        self,
        url: str,
        params: dict[str, Any] | None = None,
        extra_headers: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        """GET 请求并返回 JSON，内置重试与随机 Header。"""
        headers = _random_headers(extra_headers)
        last_exc: Exception | None = None

        for attempt in range(1, self._max_retries + 1):
            try:
                logger.debug("GET %s (attempt %d/%d)", url, attempt, self._max_retries)
                with httpx.Client(timeout=self._timeout) as client:
                    resp = client.get(url, params=params, headers=headers)
                    resp.raise_for_status()
                    return resp.json()
            except (httpx.HTTPStatusError, httpx.RequestError) as exc:
                last_exc = exc
                logger.warning(
                    "请求失败 [%d/%d]: %s — %s",
                    attempt,
                    self._max_retries,
                    url,
                    exc,
                )
                if attempt < self._max_retries:
                    _jitter_sleep(self._base_delay * attempt)
                    headers = _random_headers(extra_headers)  # 每次重试换 Header

        raise RuntimeError(f"请求 {url} 失败，已重试 {self._max_retries} 次") from last_exc

    def get_text(
        self,
        url: str,
        params: dict[str, Any] | None = None,
        extra_headers: dict[str, str] | None = None,
    ) -> str:
        """GET 请求并返回纯文本。"""
        headers = _random_headers(extra_headers)
        last_exc: Exception | None = None

        for attempt in range(1, self._max_retries + 1):
            try:
                with httpx.Client(timeout=self._timeout) as client:
                    resp = client.get(url, params=params, headers=headers)
                    resp.raise_for_status()
                    return resp.text
            except (httpx.HTTPStatusError, httpx.RequestError) as exc:
                last_exc = exc
                logger.warning("请求失败 [%d/%d]: %s — %s", attempt, self._max_retries, url, exc)
                if attempt < self._max_retries:
                    _jitter_sleep(self._base_delay * attempt)
                    headers = _random_headers(extra_headers)

        raise RuntimeError(f"请求 {url} 失败，已重试 {self._max_retries} 次") from last_exc


# 全局单例
http = HttpClient()
