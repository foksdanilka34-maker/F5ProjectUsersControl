import requests


def service_reachable(url: str, timeout: float = 2.0) -> bool:
    """Best-effort connectivity probe used to skip tests instead of failing
    them when the target service isn't running (e.g. no docker-compose stack
    on this machine)."""
    try:
        requests.get(url, timeout=timeout)
        return True
    except requests.exceptions.RequestException:
        return False
