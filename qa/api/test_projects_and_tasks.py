import uuid

import requests


def test_me_endpoint_returns_authenticated_user(gateway_url, auth_headers):
    response = requests.get(f"{gateway_url}/api/v1/auth/me", headers=auth_headers, timeout=5)
    assert response.status_code == 200

    body = response.json()
    assert body["login"]


def test_create_project_create_task_and_cleanup(gateway_url, auth_headers):
    project_name = f"qa-smoke-{uuid.uuid4().hex[:8]}"

    create_project = requests.post(
        f"{gateway_url}/api/v1/projects",
        headers=auth_headers,
        json={"name": project_name, "description": "created by the pytest smoke suite"},
        timeout=5,
    )
    assert create_project.status_code == 201
    project = create_project.json()
    project_id = project["id"]

    try:
        create_task = requests.post(
            f"{gateway_url}/api/v1/tasks",
            headers=auth_headers,
            json={"project_id": project_id, "title": "QA smoke task"},
            timeout=5,
        )
        assert create_task.status_code == 201
        task = create_task.json()
        assert task["status"] == "TODO"

        list_tasks = requests.get(
            f"{gateway_url}/api/v1/tasks",
            headers=auth_headers,
            params={"project_id": project_id},
            timeout=5,
        )
        assert list_tasks.status_code == 200
        task_ids = [t["id"] for t in list_tasks.json().get("tasks") or []]
        assert task["id"] in task_ids

        delete_task = requests.delete(
            f"{gateway_url}/api/v1/tasks/{task['id']}", headers=auth_headers, timeout=5
        )
        assert delete_task.status_code == 200
    finally:
        requests.delete(f"{gateway_url}/api/v1/projects/{project_id}", headers=auth_headers, timeout=5)


def test_unauthenticated_request_is_rejected(gateway_url):
    response = requests.get(f"{gateway_url}/api/v1/projects", timeout=5)
    assert response.status_code == 401
