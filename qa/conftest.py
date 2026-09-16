import pytest
import requests

from settings import (
    BUSINESS_URL,
    EMPLOYEE_URL,
    FRONTEND_URL,
    GATEWAY_URL,
    REQUEST_TIMEOUT,
    TEST_LOGIN,
    TEST_PASSWORD,
)
from utils import service_reachable


@pytest.fixture(scope="session")
def gateway_url():
    if not service_reachable(f"{GATEWAY_URL}/health"):
        pytest.skip(
            f"gateway is not reachable at {GATEWAY_URL} — "
            "start the stack with `make docker-up run-employee run-business run-gateway`"
        )
    return GATEWAY_URL


@pytest.fixture(scope="session")
def employee_url():
    if not service_reachable(f"{EMPLOYEE_URL}/health"):
        pytest.skip(f"employee service is not reachable at {EMPLOYEE_URL}")
    return EMPLOYEE_URL


@pytest.fixture(scope="session")
def business_url():
    if not service_reachable(f"{BUSINESS_URL}/health"):
        pytest.skip(f"business service is not reachable at {BUSINESS_URL}")
    return BUSINESS_URL


@pytest.fixture(scope="session")
def frontend_url():
    if not service_reachable(FRONTEND_URL):
        pytest.skip(
            f"frontend is not reachable at {FRONTEND_URL} — "
            "start it with `npm run dev` (or `npm run build && npm run preview`) in frontend/"
        )
    return FRONTEND_URL


@pytest.fixture(scope="session")
def auth_token(gateway_url):
    """Logs in with QA_TEST_LOGIN/QA_TEST_PASSWORD. Skips authenticated tests
    when no such user exists — this repo doesn't seed one by default."""
    response = requests.post(
        f"{gateway_url}/api/v1/auth/login",
        json={"login": TEST_LOGIN, "password": TEST_PASSWORD},
        timeout=REQUEST_TIMEOUT,
    )
    if response.status_code != 200:
        pytest.skip(
            f"could not authenticate as '{TEST_LOGIN}' (HTTP {response.status_code}) — "
            "set QA_TEST_LOGIN/QA_TEST_PASSWORD to a seeded user to run authenticated tests"
        )
    return response.json()["access_token"]


@pytest.fixture
def auth_headers(auth_token):
    return {"Authorization": f"Bearer {auth_token}"}
