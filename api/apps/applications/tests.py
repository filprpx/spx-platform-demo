from unittest.mock import patch

import pytest
from rest_framework.test import APIClient

from apps.applications.models import Application
from apps.provisioning.models import ProvisioningRequest
from apps.provisioning.publisher import ProvisioningQueueError


@pytest.fixture
def authenticated_client():
    client = APIClient()
    client.credentials(HTTP_AUTHORIZATION="Bearer test-token")
    return client


@pytest.fixture
def entra_claims():
    return {
        "oid": "user-object-id",
        "tid": "tenant-id",
        "preferred_username": "developer@example.com",
        "name": "Developer",
        "scp": "access_as_user",
    }


@pytest.mark.django_db(transaction=True)
@patch("apps.identity.authentication.EntraBearerAuthentication._decode")
@patch("apps.provisioning.services.publish_provisioning_job")
def test_create_application_records_authenticated_creator_and_queues_job(
    publish, decode, authenticated_client, entra_claims
):
    decode.return_value = entra_claims

    response = authenticated_client.post(
        "/api/v1/applications",
        {
            "name": "payments-api",
            "owning_team": "finance-engineering",
            "compute_size": "small",
            "container_port": 8080,
            "ingress": "external",
            "min_replicas": 0,
            "max_replicas": 1,
        },
        format="json",
    )

    assert response.status_code == 201
    assert response.data["name"] == "payments-api"
    assert response.data["created_by"]["email"] == "developer@example.com"
    assert response.data["provisioning_request"]["status"] == "QUEUED"
    assert Application.objects.count() == 1
    assert ProvisioningRequest.objects.count() == 1
    assert Application.objects.get().created_by.entra_object_id == "user-object-id"
    publish.assert_called_once()
    payload = publish.call_args.args[0]
    assert payload == {
        "schema_version": 2,
        "job_type": "provision_application",
        "job_id": payload["job_id"],
        "provisioning_request_id": payload["provisioning_request_id"],
        "application": {
            "id": payload["application"]["id"],
            "name": "payments-api",
            "owning_team": "finance-engineering",
            "compute_size": "small",
            "container_port": 8080,
            "ingress": "external",
            "min_replicas": 0,
            "max_replicas": 1,
        },
        "requested_by": {
            "platform_user_id": payload["requested_by"]["platform_user_id"],
            "entra_object_id": "user-object-id",
        },
    }


@pytest.mark.django_db(transaction=True)
@patch("apps.identity.authentication.EntraBearerAuthentication._decode")
@patch("apps.provisioning.services.publish_provisioning_job")
def test_duplicate_application_name_is_rejected(publish, decode, authenticated_client, entra_claims):
    decode.return_value = entra_claims
    payload = {
        "name": "payments-api",
        "owning_team": "finance",
        "compute_size": "small",
        "container_port": 8080,
        "ingress": "external",
        "min_replicas": 0,
        "max_replicas": 1,
    }

    assert authenticated_client.post("/api/v1/applications", payload, format="json").status_code == 201
    duplicate = authenticated_client.post("/api/v1/applications", payload, format="json")

    assert duplicate.status_code == 400
    assert Application.objects.count() == 1
    assert ProvisioningRequest.objects.count() == 1
    assert publish.call_count == 1


@pytest.mark.django_db(transaction=True)
@patch("apps.identity.authentication.EntraBearerAuthentication._decode")
@patch("apps.provisioning.services.publish_provisioning_job")
def test_queue_failure_returns_service_unavailable_and_preserves_request(
    publish, decode, authenticated_client, entra_claims
):
    decode.return_value = entra_claims
    publish.side_effect = ProvisioningQueueError()
    authenticated_client.raise_request_exception = False

    response = authenticated_client.post(
        "/api/v1/applications",
        {
            "name": "payments-api",
            "owning_team": "finance-engineering",
            "compute_size": "small",
            "container_port": 8080,
            "ingress": "external",
            "min_replicas": 0,
            "max_replicas": 1,
        },
        format="json",
    )

    assert response.status_code == 503
    assert ProvisioningRequest.objects.get().status == ProvisioningRequest.Status.QUEUE_FAILED


@pytest.mark.django_db
def test_application_requires_authentication():
    response = APIClient().post(
        "/api/v1/applications",
        {"name": "payments-api", "owning_team": "finance"},
        format="json",
    )

    assert response.status_code == 401
