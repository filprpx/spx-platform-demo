from unittest.mock import patch

import pytest
from rest_framework.test import APIClient

from apps.applications.models import Application
from apps.provisioning.models import ProvisioningRequest


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


@pytest.mark.django_db
@patch("apps.identity.authentication.EntraBearerAuthentication._decode")
def test_create_application_records_authenticated_creator_and_pending_request(
    decode, authenticated_client, entra_claims
):
    decode.return_value = entra_claims

    response = authenticated_client.post(
        "/api/v1/applications",
        {"name": "payments-api", "type": "api", "runtime": "go", "owning_team": "finance-engineering"},
        format="json",
    )

    assert response.status_code == 201
    assert response.data["name"] == "payments-api"
    assert response.data["created_by"]["email"] == "developer@example.com"
    assert response.data["provisioning_request"]["status"] == "PENDING"
    assert Application.objects.count() == 1
    assert ProvisioningRequest.objects.count() == 1
    assert Application.objects.get().created_by.entra_object_id == "user-object-id"


@pytest.mark.django_db
@patch("apps.identity.authentication.EntraBearerAuthentication._decode")
def test_duplicate_application_name_is_rejected(decode, authenticated_client, entra_claims):
    decode.return_value = entra_claims
    payload = {"name": "payments-api", "type": "api", "runtime": "go", "owning_team": "finance"}

    assert authenticated_client.post("/api/v1/applications", payload, format="json").status_code == 201
    duplicate = authenticated_client.post("/api/v1/applications", payload, format="json")

    assert duplicate.status_code == 400
    assert Application.objects.count() == 1
    assert ProvisioningRequest.objects.count() == 1


@pytest.mark.django_db
def test_application_requires_authentication():
    response = APIClient().post(
        "/api/v1/applications",
        {"name": "payments-api", "type": "api", "runtime": "go", "owning_team": "finance"},
        format="json",
    )

    assert response.status_code == 401
