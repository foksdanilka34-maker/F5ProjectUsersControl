# QA test suite

Pytest-based API and browser (Playwright + Selenium) smoke tests for
F5 Project & Team Control. Lives outside `backend/` and `frontend/` on
purpose — it's a separate consumer of both, not part of either.

## Setup

```bash
cd qa
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
playwright install --with-deps chromium
```

## Running

Tests skip themselves (rather than fail) when the service they need isn't
reachable, so it's safe to run `pytest` with nothing running locally —
you'll just see skips instead of a wall of connection errors.

```bash
# everything
python -m pytest -v

# just the API tests, against a stack started with:
#   make docker-up run-employee run-business run-gateway
python -m pytest api -v

# just the browser tests, against:
#   cd frontend && npm run dev
python -m pytest e2e -v
```

## Configuration

All of it is optional env vars, defaults assume the ports from the root
`Makefile` / `docker-compose.yml`:

| Variable | Default | |
| --- | --- | --- |
| `QA_GATEWAY_URL` | `http://localhost:8080` | API gateway |
| `QA_EMPLOYEE_URL` | `http://localhost:8081` | employee service (direct) |
| `QA_BUSINESS_URL` | `http://localhost:8082` | business service (direct) |
| `QA_FRONTEND_URL` | `http://localhost:5173` | Vite dev server / preview |
| `QA_TEST_LOGIN` / `QA_TEST_PASSWORD` | `admin` / `admin` | credentials for the authenticated API tests |

The repo doesn't seed a default user, so `QA_TEST_LOGIN`/`QA_TEST_PASSWORD`
won't resolve to a real account out of the box — the tests that need auth
skip with an explanation instead of failing. Point them at a real seeded
account to actually exercise that path.
