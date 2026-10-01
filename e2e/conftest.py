"""E2E test fixtures for BillionMail + FrostByte."""

import logging
import os

import httpx
import pytest
import pytest_asyncio

from helpers.api import BM_BASE, FB_BASE
from helpers.mail import MailpitClient

logging.basicConfig(level=logging.DEBUG, format="%(asctime)s %(name)s %(levelname)s %(message)s")
logger = logging.getLogger("e2e")

BM_URL = os.getenv("BM_URL", BM_BASE)
FB_URL = os.getenv("FB_URL", FB_BASE)
MAILPIT_URL = os.getenv("MAILPIT_URL", "http://localhost:8025")

# Auth token — obtained via login at session start
_auth_token: str | None = None


async def _get_auth_token(base_url: str) -> str:
    """Login to BillionMail and return JWT token."""
    global _auth_token
    if _auth_token:
        return _auth_token

    async with httpx.AsyncClient(base_url=base_url, timeout=15) as client:
        resp = await client.post(
            "/api/login",
            json={"username": "admin", "password": "admin"},
        )
        if resp.status_code == 200:
            data = resp.json()
            _auth_token = data.get("data", {}).get("token", "")
            logger.debug("[AUTH] Got token: %s...", _auth_token[:20] if _auth_token else "NONE")
        else:
            logger.warning("[AUTH] Login failed: %d %s", resp.status_code, resp.text[:200])
            _auth_token = ""
    return _auth_token


@pytest_asyncio.fixture(scope="session", loop_scope="session")
async def bm_api():
    """httpx.AsyncClient for BillionMail API."""
    token = await _get_auth_token(BM_URL)
    client = httpx.AsyncClient(
        base_url=BM_URL,
        timeout=30,
        headers={"Authorization": token},
    )
    yield client
    try:
        await client.aclose()
    except RuntimeError:
        pass  # event loop already closed


@pytest_asyncio.fixture(scope="session", loop_scope="session")
async def fb_api():
    """httpx.AsyncClient for FrostByte API."""
    client = httpx.AsyncClient(base_url=FB_URL, timeout=30)
    yield client
    try:
        await client.aclose()
    except RuntimeError:
        pass


@pytest_asyncio.fixture(scope="session", loop_scope="session")
async def mailpit():
    """Mailpit REST client."""
    client = MailpitClient(MAILPIT_URL)
    await client.clear()
    yield client


@pytest_asyncio.fixture(scope="session", loop_scope="session")
async def db():
    """Direct asyncpg connection for DB assertions."""
    import asyncpg

    db_url = os.getenv(
        "DATABASE_URL_RAW",
        "postgresql://billionmail:billionmail@localhost:25432/billionmail",
    )
    conn = await asyncpg.connect(db_url)
    yield conn
    await conn.close()


@pytest_asyncio.fixture(scope="session", loop_scope="session")
async def seed_data(bm_api, db):
    """Seed test data: template, group, contacts, API key. Returns dict of IDs."""
    from seed import seed_test_data

    data = await seed_test_data(bm_api, db)
    yield data
