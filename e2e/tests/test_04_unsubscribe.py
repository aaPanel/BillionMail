"""E2E: Unsubscribe link → contact deactivated."""

import re
import time
from urllib.parse import parse_qs, urlparse

import httpx
import pytest

from helpers import api
from helpers.api import rewrite_url
from helpers.wait import poll_until


@pytest.mark.asyncio
async def test_unsubscribe_deactivates_contact(bm_api, mailpit, db, seed_data):
    """Send email → find unsub link → follow redirect → POST unsub API → verify deactivated."""
    await mailpit.clear()

    resp = await api.post(
        bm_api,
        "/api/batch_mail/task/create",
        json={
            "addresser": seed_data["sender_email"],
            "subject": "E2E Unsub Test",
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
            subject_contains="E2E Unsub Test",
            timeout=1,
        ),
        description="unsub test email delivery",
        timeout=60,
        interval=2,
    )
    assert msg is not None

    html = await mailpit.get_message_html(msg["ID"])

    # Strategy 1: direct unsubscribe_new.html?jwt= link
    unsub_urls = re.findall(r'href="([^"]*unsubscribe_new\.html\?jwt=[^"]*)"', html, re.IGNORECASE)

    if unsub_urls:
        parsed = urlparse(unsub_urls[0])
        jwt_token = parse_qs(parsed.query).get("jwt", [None])[0]
    else:
        # Strategy 2: BM wraps unsub link in /pmta/ tracking — find by "Unsubscribe" text
        unsub_match = re.findall(r'href="([^"]+)"[^>]*>Unsubscribe</a>', html, re.IGNORECASE)
        if not unsub_match:
            # Broader fallback
            unsub_match = re.findall(r'href="([^"]*unsub[^"]*)"', html, re.IGNORECASE)
            if not unsub_match:
                unsub_match = re.findall(r'href="([^"]*pmta[^"]*)"', html, re.IGNORECASE)
        assert unsub_match, f"No unsubscribe link found in email HTML"

        # Follow the tracked unsub link — should redirect to unsubscribe_new.html?jwt=
        async with httpx.AsyncClient(timeout=10, follow_redirects=False) as client:
            redir_resp = await client.get(rewrite_url(unsub_match[0]))
            assert redir_resp.status_code in (301, 302, 307, 308), (
                f"Expected redirect from unsub tracking link, got {redir_resp.status_code}"
            )

        location = redir_resp.headers.get("location", "")
        parsed = urlparse(location)
        jwt_token = parse_qs(parsed.query).get("jwt", [None])[0]

    assert jwt_token, f"No JWT found in unsubscribe flow"

    # POST to unsubscribe API
    async with httpx.AsyncClient(timeout=10) as client:
        unsub_resp = await client.post(
            f"{str(bm_api.base_url).rstrip('/')}/api/unsubscribe_new",
            json={"jwt": jwt_token},
        )
        assert unsub_resp.status_code == 200, f"Unsubscribe POST failed: {unsub_resp.status_code}"

    # Verify contact is deactivated in DB
    row = await db.fetchrow(
        "SELECT active FROM bm_contacts WHERE email = $1 AND group_id = $2",
        seed_data["recipient_email"],
        seed_data["group_id"],
    )
    assert row is not None, "Contact not found in DB after unsubscribe"
    assert row["active"] == 0, "Contact should be deactivated after unsubscribe"
