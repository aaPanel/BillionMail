"""HTTP API helpers for BillionMail + FrostByte E2E tests."""

import logging
import re

import httpx

logger = logging.getLogger("e2e")

BM_BASE = "http://localhost:8880"
FB_BASE = "http://localhost:8001"


def rewrite_url(url: str) -> str:
    """Rewrite embedded hostname URLs to point to local BM instance.

    BM embeds tracking/unsub URLs with the configured hostname (e.g.
    https://test.billionmail.com:8443/pmta/...) which doesn't resolve in CI.
    Rewrite to BM_BASE so tests can hit the local instance.
    """
    return re.sub(r'https?://[^/]+', BM_BASE, url, count=1)


async def post(client: httpx.AsyncClient, path: str, **kwargs) -> httpx.Response:
    """POST with debug logging."""
    logger.debug("[API] POST %s%s body=%s", client.base_url, path, kwargs.get("json", ""))
    resp = await client.post(path, **kwargs)
    logger.debug("[API] Response %d: %s", resp.status_code, resp.text[:200])
    return resp


async def get(client: httpx.AsyncClient, path: str, **kwargs) -> httpx.Response:
    """GET with debug logging."""
    logger.debug("[API] GET %s%s", client.base_url, path)
    resp = await client.get(path, **kwargs)
    logger.debug("[API] Response %d: %s", resp.status_code, resp.text[:200])
    return resp


async def put(client: httpx.AsyncClient, path: str, **kwargs) -> httpx.Response:
    """PUT with debug logging."""
    logger.debug("[API] PUT %s%s", client.base_url, path)
    resp = await client.put(path, **kwargs)
    logger.debug("[API] Response %d: %s", resp.status_code, resp.text[:200])
    return resp


async def delete(client: httpx.AsyncClient, path: str, **kwargs) -> httpx.Response:
    """DELETE with debug logging."""
    logger.debug("[API] DELETE %s%s", client.base_url, path)
    resp = await client.delete(path, **kwargs)
    logger.debug("[API] Response %d: %s", resp.status_code, resp.text[:200])
    return resp
