import requests


def test_gateway_health(gateway_url):
    response = requests.get(f"{gateway_url}/health", timeout=5)
    assert response.status_code == 200

    body = response.json()
    assert body["status"] == "ok"
    assert body["gateway"] == "active"
    assert "uptime" in body


def test_employee_service_health(employee_url):
    response = requests.get(f"{employee_url}/health", timeout=5)
    assert response.status_code == 200

    body = response.json()
    assert body["status"] == "ok"
    assert body["service"] == "employee"


def test_business_service_health(business_url):
    response = requests.get(f"{business_url}/health", timeout=5)
    assert response.status_code == 200

    body = response.json()
    assert body["status"] == "ok"
    assert body["service"] == "business"


def test_gateway_returns_404_for_unknown_route(gateway_url):
    response = requests.get(f"{gateway_url}/api/v1/does-not-exist", timeout=5)
    assert response.status_code == 404
