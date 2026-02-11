"""E2E: Apify + Google Maps scraping → leads created."""

import pytest

from helpers import api
from helpers.wait import poll_until


@pytest.mark.asyncio
async def test_apify_scrape_creates_leads(fb_api, db):
    """Trigger Apify scrape → WireMock returns results → leads in DB."""
    resp = await api.post(
        fb_api,
        "/leads/apify",
        json={
            "contact_job_title": "CEO",
            "contact_city": "TestCity",
        },
    )
    assert resp.status_code == 200
    data = resp.json()
    assert "task_id" in data, f"Expected task_id in response: {data}"

    # Poll DB for leads in TestCity (Apify leads arrive as "enriched" when pre-enriched)
    async def check_leads():
        rows = await db.fetch(
            "SELECT id, business_name, status FROM frostbyte.leads WHERE city = $1",
            "TestCity",
        )
        return rows if rows else None

    leads = await poll_until(check_leads, "Apify leads created", timeout=60, interval=2)
    assert len(leads) >= 1
    assert leads[0]["business_name"] == "E2E Test Corp"
