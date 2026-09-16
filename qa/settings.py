import os

GATEWAY_URL = os.environ.get("QA_GATEWAY_URL", "http://localhost:8080").rstrip("/")
EMPLOYEE_URL = os.environ.get("QA_EMPLOYEE_URL", "http://localhost:8081").rstrip("/")
BUSINESS_URL = os.environ.get("QA_BUSINESS_URL", "http://localhost:8082").rstrip("/")
FRONTEND_URL = os.environ.get("QA_FRONTEND_URL", "http://localhost:5173").rstrip("/")

TEST_LOGIN = os.environ.get("QA_TEST_LOGIN", "admin")
TEST_PASSWORD = os.environ.get("QA_TEST_PASSWORD", "admin")

REQUEST_TIMEOUT = float(os.environ.get("QA_REQUEST_TIMEOUT", "5"))
