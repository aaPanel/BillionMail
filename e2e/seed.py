"""Seed test data for E2E tests via BillionMail API."""

import logging
import time

import httpx

from helpers import api

logger = logging.getLogger("e2e")

E2E_TEMPLATE_NAME = "E2E Test Template"
E2E_TEMPLATE_HTML = """<html><body>
<h1>Hello {{ .Subscriber.Email }}</h1>
<p>{Hi|Hey|Hello} there, this is a test email from BillionMail E2E.</p>
<p>Visit <a href="https://example.com/offer">our offer</a>.</p>
</body></html>"""

E2E_GROUP_NAME = "E2E Test Group"
E2E_CONTACT_EMAIL = "e2e-recipient@example.com"
E2E_API_NAME = "E2E Test API"
E2E_DOMAIN = "test.billionmail.com"
E2E_SENDER_LOCAL = "e2e-sender"
E2E_SENDER_PASSWORD = "e2eTest1234!"


async def seed_test_data(bm_client: httpx.AsyncClient, db=None) -> dict:
    """Create all test fixtures via BillionMail API. Returns IDs dict."""
    data = {}

    # 0. Create domain + sender mailbox (required for SMTP auth)
    resp = await api.post(
        bm_client,
        "/api/domains/create",
        json={"domain": E2E_DOMAIN, "quota": 1073741824, "mailboxes": 50, "rateLimit": 100},
    )
    if resp.status_code == 200 and resp.json().get("code") == 0:
        logger.info("[SEED] Created domain %s", E2E_DOMAIN)
    else:
        logger.info("[SEED] Domain may already exist: %s", resp.text[:200])

    resp = await api.post(
        bm_client,
        "/api/mailbox/create",
        json={
            "domain": E2E_DOMAIN,
            "local_part": E2E_SENDER_LOCAL,
            "password": E2E_SENDER_PASSWORD,
            "full_name": "E2E Sender",
            "active": 1,
            "isAdmin": 0,
        },
    )
    if resp.status_code == 200 and resp.json().get("code") == 0:
        logger.info("[SEED] Created mailbox %s@%s", E2E_SENDER_LOCAL, E2E_DOMAIN)
    else:
        logger.info("[SEED] Mailbox may already exist: %s", resp.text[:200])

    # 1. Create email template
    resp = await api.post(
        bm_client,
        "/api/email_template/create",
        json={
            "temp_name": E2E_TEMPLATE_NAME,
            "html_content": E2E_TEMPLATE_HTML,
            "add_type": 0,
        },
    )
    if resp.status_code == 200 and resp.json().get("code") == 0:
        data["template_id"] = resp.json().get("data", {}).get("id")
        logger.info("[SEED] Created template id=%s", data.get("template_id"))
    else:
        # Template may already exist — list and find it
        resp = await api.get(bm_client, "/api/email_template/list", params={"page": 1, "page_size": 100})
        for t in resp.json().get("data", {}).get("list", []):
            if t.get("temp_name") == E2E_TEMPLATE_NAME:
                data["template_id"] = t["id"]
                break
        logger.info("[SEED] Using existing template id=%s", data.get("template_id"))

    # 2. Create contact group (API returns no ID — must list to find it)
    resp = await api.post(
        bm_client,
        "/api/contact/group/create",
        json={"name": E2E_GROUP_NAME, "description": "E2E testing group", "create_type": 1},
    )
    if resp.status_code == 200 and resp.json().get("code") == 0:
        logger.info("[SEED] Group create succeeded, fetching ID from list")
    else:
        logger.info("[SEED] Group may already exist, looking up by name")

    resp = await api.get(bm_client, "/api/contact/group/all")
    group_data = resp.json().get("data", {})
    group_list = group_data.get("list", group_data) if isinstance(group_data, dict) else group_data
    for g in (group_list or []):
        if g.get("name") == E2E_GROUP_NAME:
            data["group_id"] = g["id"]
            break
    logger.info("[SEED] Group id=%s", data.get("group_id"))

    # 3. Import contacts into group
    if data.get("group_id"):
        contacts_str = (
            'e2e-recipient@example.com,{"first_name":"E2E","last_name":"Tester"}\n'
            'e2e-second@example.com,{"first_name":"Second","last_name":"User"}'
        )
        resp = await api.post(
            bm_client,
            "/api/contact/group/import",
            json={
                "group_ids": [data["group_id"]],
                "contacts": contacts_str,
                "import_type": 2,
                "default_active": 1,
                "status": 1,
            },
        )
        logger.info("[SEED] Imported contacts: %d", resp.status_code)

    # 3b. Re-activate contacts (may have been deactivated by previous test runs)
    if db and data.get("group_id"):
        await db.execute(
            "UPDATE bm_contacts SET active = 1 WHERE group_id = $1",
            data["group_id"],
        )
        logger.info("[SEED] Re-activated all contacts in group %s", data["group_id"])

    # 4. Create API template (for programmatic send)
    if data.get("template_id") and data.get("group_id"):
        resp = await api.post(
            bm_client,
            "/api/batch_mail/api/create",
            json={
                "api_name": E2E_API_NAME,
                "template_id": data["template_id"],
                "group_id": data["group_id"],
                "subject": "E2E API Test: {{ .Subscriber.Email }}",
                "addresser": "e2e-sender@test.billionmail.com",
                "full_name": "E2E Sender",
                "unsubscribe": 1,
                "active": 1,
                "expire_time": 0,
                "track_open": 1,
                "track_click": 1,
            },
        )
        if resp.status_code == 200:
            # The API key is auto-generated; fetch it from list
            list_resp = await api.get(
                bm_client,
                "/api/batch_mail/api/list",
                params={"page": 1, "page_size": 50, "keyword": E2E_API_NAME},
            )
            api_data = list_resp.json().get("data") or {}
            for item in (api_data.get("list") or []):
                if item.get("api_name") == E2E_API_NAME:
                    data["api_key"] = item.get("api_key")
                    data["api_id"] = item.get("id")
                    break
            logger.info("[SEED] Created API template, key=%s", data.get("api_key", "")[:10])

    data["sender_email"] = "e2e-sender@test.billionmail.com"
    data["recipient_email"] = E2E_CONTACT_EMAIL
    data["timestamp"] = int(time.time())

    logger.info("[SEED] Seed complete: %s", {k: v for k, v in data.items() if k != "api_key"})
    return data
