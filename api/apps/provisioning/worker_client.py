import json
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

from django.conf import settings


class CallbackDeliveryError(RuntimeError):
    """Raised when the worker cannot deliver a status event to Django."""


def post_execution_event(*, job_id, provisioning_request_id, status, message):
    token = settings.WORKER_CALLBACK_TOKEN
    if not token:
        raise CallbackDeliveryError("WORKER_CALLBACK_TOKEN is not configured")
    body = json.dumps(
        {
            "schema_version": 2,
            "job_id": job_id,
            "provisioning_request_id": provisioning_request_id,
            "status": status,
            "message": message,
        }
    ).encode("utf-8")
    request = Request(
        f"{settings.PLATFORM_API_URL.rstrip('/')}/internal/execution-events",
        data=body,
        headers={"Content-Type": "application/json", "X-Worker-Token": token},
        method="POST",
    )
    try:
        with urlopen(request, timeout=10) as response:
            if response.status >= 400:
                raise CallbackDeliveryError(f"callback returned HTTP {response.status}")
    except (HTTPError, URLError, TimeoutError, OSError) as exc:
        raise CallbackDeliveryError("execution event callback failed") from exc
