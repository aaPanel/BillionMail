"""E2E: Programmatic send via API key → delivery + attrib rendering verification."""

import pytest

from helpers import api
from helpers.wait import poll_until


@pytest.mark.asyncio
async def test_api_mail_send(bm_api, mailpit, seed_data):
    """POST to API send endpoint with x-api-key → verify email in Mailpit + attribs rendered."""
    if not seed_data.get("api_key"):
        pytest.skip("No API key seeded")

    await mailpit.clear()

    resp = await api.post(
        bm_api,
        "/api/batch_mail/api/send",
        headers={"x-api-key": seed_data["api_key"]},
        json={
            "recipient": seed_data["recipient_email"],
            "attribs": {"first_name": "E2E", "offer": "50% off"},
        },
    )
    assert resp.status_code == 200
    assert resp.json().get("code") == 0, f"API send failed: {resp.json()}"

    # Poll Mailpit for delivery
    msg = await poll_until(
        lambda: mailpit.wait_for_email(
            to=seed_data["recipient_email"],
            subject_contains="E2E API Test",
            timeout=1,
        ),
        description="API mail delivery",
        timeout=60,
        interval=2,
    )
    assert msg is not None, "API-sent email not received in Mailpit"

    # Verify attrib values are rendered (not raw template syntax)
    html = await mailpit.get_message_html(msg["ID"])
    assert "{{ .Subscriber" not in html, "Template syntax not rendered in HTML"
    assert "{{.Subscriber" not in html, "Template syntax not rendered in HTML"


@pytest.mark.asyncio
async def test_api_mail_invalid_key(bm_api, seed_data):
    """Send with invalid API key → expect rejection."""
    resp = await api.post(
        bm_api,
        "/api/batch_mail/api/send",
        headers={"x-api-key": "invalid-key-xxx"},
        json={
            "recipient": seed_data["recipient_email"],
            "attribs": {},
        },
    )
    # Should fail auth
    assert resp.status_code != 200 or resp.json().get("code") != 0
