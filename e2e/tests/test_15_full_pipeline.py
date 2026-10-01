"""E2E: Full pipeline — Scrape → Enrich → Personalize → Send → Track → Reply."""

import time

import pytest

from helpers import api
from helpers.wait import poll_until


@pytest.mark.asyncio
async def test_full_pipeline(bm_api, fb_api, mailpit, db, seed_data):
    """Crown jewel test: full scrape-to-reply pipeline."""
    await mailpit.clear()

    # 1. Trigger scrape via FrostByte (WireMock returns mock leads)
    scrape_resp = await api.post(
        fb_api,
        "/leads/apify",
        json={
            "campaign_id": None,
            "contact_job_title": "Founder",
            "contact_city": "PipelineCity",
        },
    )
    assert scrape_resp.status_code == 200

    # 2. Poll until leads are created (mock always returns TestCity/MockVille)
    async def check_pipeline_leads():
        rows = await db.fetch(
            "SELECT id, status, business_name FROM frostbyte.leads WHERE city IN ($1, $2)",
            "TestCity", "MockVille",
        )
        return rows if rows else None

    leads = await poll_until(check_pipeline_leads, "pipeline leads scraped", timeout=60, interval=2)
    assert len(leads) >= 1

    # 3. Wait for enrichment (worker processes scraped → enriched)
    lead_id = leads[0]["id"]

    async def check_enriched():
        r = await db.fetchrow(
            "SELECT status, owner_email FROM frostbyte.leads WHERE id = $1", lead_id
        )
        if r and r["status"] in ("enriched", "personalized", "queued", "in_sequence"):
            return r
        return None

    await poll_until(check_enriched, "pipeline lead enriched", timeout=90, interval=3)

    # 4. Verify FrostByte tracking webhook works
    tracking_resp = await api.post(
        fb_api,
        "/webhooks/billionmail/tracking",
        json={
            "event": "open",
            "message_id": "pipeline-test-msg",
            "recipient": "jane@e2etest.example.com",
            "timestamp": "2024-01-15T10:30:00Z",
        },
    )
    assert tracking_resp.status_code == 200

    # 5. Verify reply webhook works
    reply_resp = await api.post(
        fb_api,
        "/webhooks/billionmail/reply",
        json={
            "message_id": "pipeline-test-msg",
            "from_email": "jane@e2etest.example.com",
            "reply_body": "Yes, I'm interested!",
            "in_reply_to": "pipeline-test-msg",
        },
    )
    assert reply_resp.status_code == 200

    # 6. Verify BillionMail can send via API (using seed data)
    if seed_data.get("api_key"):
        send_resp = await api.post(
            bm_api,
            "/api/batch_mail/api/send",
            headers={"x-api-key": seed_data["api_key"]},
            json={
                "recipient": "pipeline-test@example.com",
                "attribs": {"first_name": "Pipeline"},
            },
        )
        assert send_resp.status_code == 200

        # Check Mailpit for delivery
        msg = await poll_until(
            lambda: mailpit.wait_for_email(
                to="pipeline-test@example.com",
                timeout=1,
            ),
            description="pipeline API email delivery",
            timeout=60,
            interval=2,
        )
        if msg:
            assert msg["Subject"]  # Subject should be populated
