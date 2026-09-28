import json

from celery import current_app
from rest_framework.exceptions import APIException


TASK_NAME = "provisioning.execute"
QUEUE_NAME = "provisioning"


class ProvisioningQueueError(APIException):
    status_code = 503
    default_detail = "The provisioning queue is unavailable."
    default_code = "provisioning_queue_unavailable"


def publish_provisioning_job(payload: dict) -> str:
    """Publish one JSON provisioning job and return its Celery task ID."""
    try:
        json.dumps(payload)
        result = current_app.send_task(
            TASK_NAME,
            args=[payload],
            queue=QUEUE_NAME,
            serializer="json",
        )
    except (TypeError, ValueError) as exc:
        raise ProvisioningQueueError("The provisioning job is not JSON serializable.") from exc
    except Exception as exc:
        raise ProvisioningQueueError() from exc
    return result.id
