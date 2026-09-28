from unittest.mock import Mock, patch
from uuid import uuid4

import pytest
from django.db import transaction
from django.test import override_settings
from rest_framework.test import APIClient
from rest_framework.exceptions import APIException

from apps.applications.models import Application
from apps.identity.models import PlatformUser
from apps.provisioning.job_contract import build_provisioning_job
from apps.provisioning.desired_state import DesiredStateError, write_desired_state
from apps.provisioning.models import ProvisioningRequest
from apps.provisioning.publisher import QUEUE_NAME, TASK_NAME, ProvisioningQueueError, publish_provisioning_job
from apps.provisioning.services import create_application_with_provisioning_request
from apps.provisioning.tasks import execute_provisioning_job
from apps.provisioning.contracts import ProvisioningJob


@pytest.fixture
def platform_user(db):
    return PlatformUser.objects.create(
        entra_object_id="user-object-id",
        tenant_id="tenant-id",
        email="developer@example.com",
        display_name="Developer",
    )


@pytest.fixture
def application_data():
    return {
        "name": "payments-api",
        "owning_team": "finance-engineering",
        "compute_size": "small",
        "container_port": 8080,
        "ingress": "external",
        "min_replicas": 0,
        "max_replicas": 1,
    }


@pytest.mark.django_db
def test_job_is_not_published_before_transaction_commit(platform_user, application_data):
    publisher = Mock()
    with pytest.raises(RuntimeError):
        with pytest.MonkeyPatch.context() as monkeypatch:
            monkeypatch.setattr(
                "apps.provisioning.services.publish_provisioning_job",
                publisher,
            )
            with transaction.atomic():
                create_application_with_provisioning_request(
                    validated_data=application_data,
                    user=platform_user,
                    publisher=publisher,
                )
                assert publisher.call_count == 0
                raise RuntimeError("rollback")
    assert publisher.call_count == 0
    assert Application.objects.count() == 0
    assert ProvisioningRequest.objects.count() == 0


@pytest.mark.django_db
def test_queue_failure_is_recorded_and_returned_as_service_unavailable(
    platform_user, application_data, django_capture_on_commit_callbacks
):
    def fail(_payload):
        raise ProvisioningQueueError()

    with pytest.raises(APIException) as raised:
        with django_capture_on_commit_callbacks(execute=True):
            create_application_with_provisioning_request(
                validated_data=application_data,
                user=platform_user,
                publisher=fail,
            )

    request = ProvisioningRequest.objects.get()
    assert request.status == ProvisioningRequest.Status.QUEUE_FAILED
    assert raised.value.status_code == 503


@pytest.mark.django_db
def test_job_contract_uses_persisted_values(platform_user, application_data):
    application = Application.objects.create(created_by=platform_user, **application_data)
    request = ProvisioningRequest.objects.create(
        application=application, requested_by=platform_user, status=ProvisioningRequest.Status.QUEUED
    )

    application.name = "changed-after-payload"
    application.save(update_fields=["name"])

    payload = build_provisioning_job(application, request)
    assert payload["application"]["name"] == "changed-after-payload"
    assert "DJANGO_SECRET_KEY" not in str(payload)
    assert "access_token" not in str(payload)


def test_publisher_uses_provisioning_queue_and_json_serializer():
    result = Mock(id="celery-task-id")
    with patch("apps.provisioning.publisher.current_app.send_task", return_value=result) as send_task:
        task_id = publish_provisioning_job({"schema_version": 2, "job_type": "provision_application"})

    assert task_id == "celery-task-id"
    send_task.assert_called_once_with(
        TASK_NAME,
        args=[{"schema_version": 2, "job_type": "provision_application"}],
        queue=QUEUE_NAME,
        serializer="json",
    )


def worker_payload():
    return {
        "schema_version": 2,
        "job_type": "provision_application",
        "job_id": "11111111-1111-4111-8111-111111111111",
        "provisioning_request_id": "11111111-1111-4111-8111-111111111111",
        "application": {
            "id": "22222222-2222-4222-8222-222222222222",
            "name": "payments-api",
            "owning_team": "finance-engineering",
            "compute_size": "small",
            "container_port": 8080,
            "ingress": "external",
            "min_replicas": 0,
            "max_replicas": 1,
        },
        "requested_by": {
            "platform_user_id": "33333333-3333-4333-8333-333333333333",
            "entra_object_id": "entra-object-id",
        },
    }


def test_worker_task_reports_running_and_ready():
    with patch("apps.provisioning.tasks.post_execution_event") as callback, patch(
        "apps.provisioning.tasks.write_desired_state"
    ):
        result = execute_provisioning_job.run(worker_payload())

    assert result["status"] == "WAITING_FOR_APPROVAL"
    assert [call.kwargs["status"] for call in callback.call_args_list] == [
        "RUNNING",
        "READY_FOR_EXECUTION",
        "WAITING_FOR_APPROVAL",
    ]


def test_worker_task_reports_invalid_job_without_retrying():
    payload = worker_payload()
    payload["job_type"] = "not-supported"
    with patch("apps.provisioning.tasks.post_execution_event") as callback:
        result = execute_provisioning_job.run(payload)

    assert result["status"] == "WORKER_FAILED"
    callback.assert_called_once()
    assert callback.call_args.kwargs["status"] == "WORKER_FAILED"


def test_desired_state_is_deterministic_and_safe(tmp_path, monkeypatch):
    monkeypatch.setenv("PLATFORM_ACR_NAME", "spxdemo12345678")
    monkeypatch.setenv("PLATFORM_ACR_LOGIN_SERVER", "spxdemo12345678.azurecr.io")
    job = ProvisioningJob.model_validate(worker_payload())
    workload = write_desired_state(job, tmp_path)

    assert workload == tmp_path / "infra/workloads/dev/teams/finance-engineering/payments-api"
    assert (workload / "main.tf").exists()
    assert (workload / "terraform.tfvars").exists()
    assert (tmp_path / ".local/spx/pending/11111111-1111-4111-8111-111111111111.json").exists()
    with pytest.raises(DesiredStateError):
        write_desired_state(job, tmp_path)


def test_desired_state_rejects_unsafe_names(tmp_path):
    payload = worker_payload()
    payload["application"]["name"] = "../outside"
    job = ProvisioningJob.model_validate(payload)
    with pytest.raises(DesiredStateError):
        write_desired_state(job, tmp_path)


def test_worker_task_is_late_acknowledged_and_bounded():
    assert execute_provisioning_job.name == "provisioning.execute"
    assert execute_provisioning_job.acks_late is True
    assert execute_provisioning_job.reject_on_worker_lost is True
    assert execute_provisioning_job.retry_kwargs["max_retries"] == 3


@pytest.mark.django_db
@override_settings(WORKER_CALLBACK_TOKEN="worker-secret")
def test_execution_event_updates_request_and_is_idempotent(platform_user, application_data):
    application = Application.objects.create(created_by=platform_user, **application_data)
    request = ProvisioningRequest.objects.create(
        application=application, requested_by=platform_user, status=ProvisioningRequest.Status.QUEUED
    )
    client = APIClient()
    event = {
        "schema_version": 2,
        "job_id": str(request.id),
        "provisioning_request_id": str(request.id),
        "status": "RUNNING",
        "message": "Worker accepted provisioning job",
    }

    response = client.post("/internal/execution-events", event, format="json", HTTP_X_WORKER_TOKEN="worker-secret")
    assert response.status_code == 200
    request.refresh_from_db()
    assert request.status == ProvisioningRequest.Status.RUNNING
    assert request.started_at is not None

    duplicate = client.post("/internal/execution-events", event, format="json", HTTP_X_WORKER_TOKEN="worker-secret")
    assert duplicate.status_code == 200

    event["status"] = "READY_FOR_EXECUTION"
    response = client.post("/internal/execution-events", event, format="json", HTTP_X_WORKER_TOKEN="worker-secret")
    assert response.status_code == 200
    request.refresh_from_db()
    assert request.status == ProvisioningRequest.Status.READY_FOR_EXECUTION


@pytest.mark.django_db
@override_settings(WORKER_CALLBACK_TOKEN="worker-secret")
def test_execution_event_rejects_bad_token_and_regression(platform_user, application_data):
    application = Application.objects.create(created_by=platform_user, **application_data)
    request = ProvisioningRequest.objects.create(
        application=application, requested_by=platform_user, status=ProvisioningRequest.Status.RUNNING
    )
    client = APIClient()
    event = {
        "schema_version": 2,
        "job_id": str(request.id),
        "provisioning_request_id": str(request.id),
        "status": "READY_FOR_EXECUTION",
    }
    assert client.post("/internal/execution-events", event, format="json").status_code == 401
    event["status"] = "RUNNING"
    assert client.post("/internal/execution-events", event, format="json", HTTP_X_WORKER_TOKEN="worker-secret").status_code == 200
    event["status"] = "READY_FOR_EXECUTION"
    assert client.post("/internal/execution-events", event, format="json", HTTP_X_WORKER_TOKEN="worker-secret").status_code == 200
    event["status"] = "RUNNING"
    assert client.post("/internal/execution-events", event, format="json", HTTP_X_WORKER_TOKEN="worker-secret").status_code == 409
    event["status"] = "WORKER_FAILED"
    event["job_id"] = str(uuid4())
    assert client.post("/internal/execution-events", event, format="json", HTTP_X_WORKER_TOKEN="worker-secret").status_code == 409
