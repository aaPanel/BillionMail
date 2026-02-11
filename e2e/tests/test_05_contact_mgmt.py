"""E2E: Contact management — import, groups, tag filtering."""

import time

import pytest

from helpers import api


async def _create_group(bm_api, name, description=""):
    """Create group and return its ID. Handles 'already exists' gracefully."""
    resp = await api.post(
        bm_api,
        "/api/contact/group/create",
        json={"name": name, "description": description, "create_type": 1},
    )
    assert resp.status_code == 200

    # List groups to find the group by name (API returns no ID on create)
    resp = await api.get(bm_api, "/api/contact/group/all")
    for g in resp.json().get("data", {}).get("list", resp.json().get("data", [])):
        if g.get("name") == name:
            return g["id"]
    raise AssertionError(f"Group '{name}' not found in group list")


async def _find_tag_id(bm_api, group_id, tag_name):
    """List tags for group and find by name."""
    resp = await api.get(
        bm_api, "/api/tags/list",
        params={"group_id": group_id, "keyword": tag_name, "page": 1, "page_size": 50},
    )
    for t in resp.json().get("data", {}).get("list", []):
        if t.get("name") == tag_name:
            return t["id"]
    return None


@pytest.mark.asyncio
async def test_create_group_and_import_contacts(bm_api):
    """Create group → import contacts via paste format → verify count."""
    ts = int(time.time())
    group_id = await _create_group(bm_api, f"E2E Contact Test {ts}", "Testing contact mgmt")

    # Import contacts using correct API contract
    contacts_str = (
        'contact-a@e2e.test,{"first_name":"Alice"}\n'
        'contact-b@e2e.test,{"first_name":"Bob"}\n'
        'contact-c@e2e.test,{"first_name":"Charlie"}'
    )
    resp = await api.post(
        bm_api,
        "/api/contact/group/import",
        json={
            "group_ids": [group_id],
            "contacts": contacts_str,
            "import_type": 2,
            "default_active": 1,
            "status": 1,
        },
    )
    assert resp.status_code == 200

    # List contacts and verify
    resp = await api.get(
        bm_api,
        "/api/contact/list",
        params={"group_id": group_id, "page": 1, "page_size": 50},
    )
    assert resp.status_code == 200
    contacts = resp.json().get("data", {}).get("list", [])
    emails = [c["email"] for c in contacts]
    assert "contact-a@e2e.test" in emails
    assert "contact-b@e2e.test" in emails
    assert "contact-c@e2e.test" in emails

    # Cleanup
    await api.post(bm_api, "/api/contact/group/delete", json={"group_ids": [group_id]})


@pytest.mark.asyncio
async def test_list_all_groups(bm_api, seed_data):
    """Verify seed group appears in group list."""
    resp = await api.get(bm_api, "/api/contact/group/all")
    assert resp.status_code == 200
    groups = resp.json().get("data", {}).get("list", [])
    group_ids = [g["id"] for g in groups]
    assert seed_data["group_id"] in group_ids


@pytest.mark.asyncio
async def test_tag_filtering_and_or_not(bm_api):
    """Create tags → assign to contacts → verify tag_contact_count with AND/OR/NOT."""
    ts = int(time.time())
    group_id = await _create_group(bm_api, f"E2E Tag Test {ts}", "Tag filtering test")

    contacts_str = (
        'tag-a@e2e.test,{"first_name":"Alice"}\n'
        'tag-b@e2e.test,{"first_name":"Bob"}\n'
        'tag-c@e2e.test,{"first_name":"Charlie"}'
    )
    resp = await api.post(
        bm_api,
        "/api/contact/group/import",
        json={
            "group_ids": [group_id],
            "contacts": contacts_str,
            "import_type": 2,
            "default_active": 1,
            "status": 1,
        },
    )
    assert resp.status_code == 200
    assert resp.json().get("code") == 0, f"Import failed: {resp.json()}"

    # Get contact IDs
    resp = await api.get(
        bm_api, "/api/contact/list",
        params={"group_id": group_id, "page": 1, "page_size": 50},
    )
    contacts = resp.json().get("data", {}).get("list", [])
    contact_map = {c["email"]: c["id"] for c in contacts}
    id_a = contact_map.get("tag-a@e2e.test")
    id_b = contact_map.get("tag-b@e2e.test")
    id_c = contact_map.get("tag-c@e2e.test")
    assert id_a and id_b and id_c, f"Missing contacts: {contact_map}"

    # Create 2 tags (requires group_id)
    tag1_name = f"e2e-tag-alpha-{ts}"
    resp = await api.post(bm_api, "/api/tags/create", json={"name": tag1_name, "group_id": group_id})
    assert resp.status_code == 200
    assert resp.json().get("code") == 0, f"Tag1 creation failed: {resp.json()}"
    tag1_id = await _find_tag_id(bm_api, group_id, tag1_name)
    assert tag1_id, f"Tag1 '{tag1_name}' not found in list"

    tag2_name = f"e2e-tag-beta-{ts}"
    resp = await api.post(bm_api, "/api/tags/create", json={"name": tag2_name, "group_id": group_id})
    assert resp.status_code == 200
    assert resp.json().get("code") == 0, f"Tag2 creation failed: {resp.json()}"
    tag2_id = await _find_tag_id(bm_api, group_id, tag2_name)
    assert tag2_id, f"Tag2 '{tag2_name}' not found in list"

    # Assign tag1 to A+B (action=1 is Add)
    resp = await api.post(
        bm_api, "/api/contact/batch_tags_opt",
        json={"ids": [id_a, id_b], "tag_ids": [tag1_id], "action": 1},
    )
    assert resp.status_code == 200
    assert resp.json().get("code") == 0, f"Batch tag1 failed: {resp.json()}"

    # Assign tag2 to B+C
    resp = await api.post(
        bm_api, "/api/contact/batch_tags_opt",
        json={"ids": [id_b, id_c], "tag_ids": [tag2_id], "action": 1},
    )
    assert resp.status_code == 200
    assert resp.json().get("code") == 0, f"Batch tag2 failed: {resp.json()}"

    # Verify AND (tag1+tag2) → total=1 (only B has both)
    resp = await api.post(
        bm_api, "/api/contact/group/tag_contact_count",
        json={"group_id": group_id, "tag_ids": [tag1_id, tag2_id], "tag_logic": "AND"},
    )
    assert resp.status_code == 200
    and_total = resp.json().get("data", {}).get("total", -1)
    assert and_total == 1, f"AND total expected 1, got {and_total}: {resp.json()}"

    # Verify OR (tag1|tag2) → total=3 (A,B,C)
    resp = await api.post(
        bm_api, "/api/contact/group/tag_contact_count",
        json={"group_id": group_id, "tag_ids": [tag1_id, tag2_id], "tag_logic": "OR"},
    )
    assert resp.status_code == 200
    or_total = resp.json().get("data", {}).get("total", -1)
    assert or_total == 3, f"OR total expected 3, got {or_total}: {resp.json()}"

    # Verify NOT (tag1) → total=1 (only C doesn't have tag1)
    resp = await api.post(
        bm_api, "/api/contact/group/tag_contact_count",
        json={"group_id": group_id, "tag_ids": [tag1_id], "tag_logic": "NOT"},
    )
    assert resp.status_code == 200
    not_total = resp.json().get("data", {}).get("total", -1)
    assert not_total == 1, f"NOT total expected 1, got {not_total}: {resp.json()}"

    # Cleanup
    await api.post(bm_api, "/api/tags/delete", json={"id": tag1_id})
    await api.post(bm_api, "/api/tags/delete", json={"id": tag2_id})
    await api.post(bm_api, "/api/contact/group/delete", json={"group_ids": [group_id]})
