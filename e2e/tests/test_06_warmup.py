"""E2E: Warmup campaign creation and DB verification."""

import time

import pytest

from helpers import api


@pytest.mark.asyncio
async def test_warmup_campaign_creation(bm_api, db, seed_data):
    """Create campaign with warmup=1 → verify task created + DB record exists."""
    resp = await api.post(
        bm_api,
        "/api/batch_mail/task/create",
        json={
            "addresser": seed_data["sender_email"],
            "subject": "E2E Warmup Test",
            "full_name": "E2E Sender",
            "template_id": seed_data["template_id"],
            "group_id": seed_data["group_id"],
            "start_time": int(time.time()) + 3600,  # future start to avoid sending
            "track_open": 1,
            "track_click": 1,
            "unsubscribe": 1,
            "threads": 1,
            "warmup": 1,
        },
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body.get("code") == 0, f"Warmup task creation failed: {body}"
    task_id = body["data"]["id"]

    # Verify task has warmup flag via API
    info_resp = await api.get(bm_api, "/api/batch_mail/task/find", params={"id": task_id})
    assert info_resp.status_code == 200
    task_data = info_resp.json().get("data", {})
    assert task_data, f"Task info returned empty data for id={task_id}"

    # Verify warmup record in DB
    row = await db.fetchrow(
        "SELECT id FROM bm_campaign_warmup WHERE task_id = $1",
        task_id,
    )
    assert row is not None, f"No bm_campaign_warmup record for task_id={task_id}"

    # Cleanup — delete task
    await api.post(bm_api, "/api/batch_mail/task/delete", json={"id": task_id})
