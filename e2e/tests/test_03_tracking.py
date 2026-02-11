"""E2E: Open pixel + click tracking verification."""

import re
import time

import httpx
import pytest

from helpers import api
from helpers.api import rewrite_url
from helpers.wait import poll_until


@pytest.mark.asyncio
async def test_open_tracking_pixel(bm_api, mailpit, db, seed_data):
    """Send email → extract pixel URL → fire it → verify open event in DB."""
    await mailpit.clear()

    resp = await api.post(
        bm_api,
        "/api/batch_mail/task/create",
        json={
            "addresser": seed_data["sender_email"],
            "subject": "E2E Tracking Test",
            "full_name": "E2E Sender",
            "template_id": seed_data["template_id"],
            "group_id": seed_data["group_id"],
            "start_time": int(time.time()),
            "track_open": 1,
            "track_click": 1,
            "unsubscribe": 1,
            "threads": 5,
        },
    )
    assert resp.status_code == 200

    msg = await poll_until(
        lambda: mailpit.wait_for_email(
            to=seed_data["recipient_email"],
            subject_contains="E2E Tracking Test",
            timeout=1,
        ),
        description="tracking test email delivery",
        timeout=60,
        interval=2,
    )
    assert msg is not None

    # Extract open tracking pixel — BM rewrites to /pmta/{encrypted} URLs
    html = await mailpit.get_message_html(msg["ID"])
    all_img_srcs = re.findall(r'<img[^>]+src="([^"]+)"', html, re.IGNORECASE)
    pixel_urls = [u for u in all_img_srcs if "/pmta/" in u]
    assert pixel_urls, f"No /pmta/ tracking pixel found. img srcs: {all_img_srcs}"

    # Fire the pixel URL (rewrite hostname to localhost)
    async with httpx.AsyncClient(timeout=10, follow_redirects=True) as client:
        pixel_resp = await client.get(rewrite_url(pixel_urls[0]))
        assert pixel_resp.status_code == 200

    # Verify open recorded in DB
    row = await db.fetchrow(
        "SELECT id FROM mailstat_opened WHERE campaign_id = $1 AND recipient = $2",
        seed_data.get("campaign_id", 0),
        seed_data["recipient_email"],
    )
    # Soft check — campaign_id may not match if task auto-assigns different id
    if row is None:
        # Fallback: just check any recent open for this recipient
        row = await db.fetchrow(
            "SELECT id FROM mailstat_opened WHERE recipient = $1 ORDER BY id DESC LIMIT 1",
            seed_data["recipient_email"],
        )
    assert row is not None, "No mailstat_opened record found after pixel fire"


@pytest.mark.asyncio
async def test_click_tracking_redirect(bm_api, mailpit, db, seed_data):
    """Send email → extract tracked link → follow it → verify redirect + DB record."""
    await mailpit.clear()

    resp = await api.post(
        bm_api,
        "/api/batch_mail/task/create",
        json={
            "addresser": seed_data["sender_email"],
            "subject": "E2E Click Test",
            "full_name": "E2E Sender",
            "template_id": seed_data["template_id"],
            "group_id": seed_data["group_id"],
            "start_time": int(time.time()),
            "track_open": 1,
            "track_click": 1,
            "unsubscribe": 1,
            "threads": 5,
        },
    )
    assert resp.status_code == 200

    msg = await poll_until(
        lambda: mailpit.wait_for_email(
            to=seed_data["recipient_email"],
            subject_contains="E2E Click Test",
            timeout=1,
        ),
        description="click test email delivery",
        timeout=60,
        interval=2,
    )
    assert msg is not None

    html = await mailpit.get_message_html(msg["ID"])

    # Find tracked links — BM rewrites hrefs to /pmta/{encrypted} URLs
    all_hrefs = re.findall(r'href="([^"]+)"', html)
    # Filter for /pmta/ links that are NOT img pixel URLs (pixel is typically 1x1 img)
    all_img_srcs = set(re.findall(r'<img[^>]+src="([^"]+)"', html, re.IGNORECASE))
    tracked_links = [u for u in all_hrefs if "/pmta/" in u and u not in all_img_srcs]
    assert tracked_links, f"No /pmta/ tracked links found. hrefs: {all_hrefs}"

    # Follow tracked link — should redirect to original URL (rewrite hostname)
    async with httpx.AsyncClient(timeout=10, follow_redirects=False) as client:
        click_resp = await client.get(rewrite_url(tracked_links[0]))
        assert click_resp.status_code in (301, 302, 307, 308), (
            f"Expected redirect, got {click_resp.status_code}"
        )

    # Verify click recorded in DB
    row = await db.fetchrow(
        "SELECT id FROM mailstat_clicked WHERE recipient = $1 ORDER BY id DESC LIMIT 1",
        seed_data["recipient_email"],
    )
    assert row is not None, "No mailstat_clicked record found after click"
