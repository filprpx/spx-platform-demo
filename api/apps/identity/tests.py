from unittest.mock import patch
from datetime import datetime, timedelta, timezone
from types import SimpleNamespace

import jwt
import pytest
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa
from django.test import override_settings
from rest_framework.exceptions import AuthenticationFailed
from rest_framework.request import Request
from rest_framework.test import APIRequestFactory, APIClient

from .authentication import EntraBearerAuthentication
from .models import PlatformUser


@pytest.mark.django_db
@patch("apps.identity.authentication.EntraBearerAuthentication._decode")
def test_me_is_idempotent_for_the_same_entra_identity(decode):
    decode.return_value = {
        "oid": "user-object-id",
        "tid": "tenant-id",
        "preferred_username": "developer@example.com",
        "name": "Developer",
        "scp": "access_as_user",
    }
    client = APIClient()
    client.credentials(HTTP_AUTHORIZATION="Bearer test-token")

    first = client.get("/api/v1/me")
    second = client.get("/api/v1/me")

    assert first.status_code == 200
    assert second.status_code == 200
    assert first.data["id"] == second.data["id"]
    assert PlatformUser.objects.count() == 1


@pytest.fixture
def signed_token():
    private_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
    public_key = private_key.public_key()
    now = datetime.now(timezone.utc)
    claims = {
        "oid": "jwt-user-object-id",
        "tid": "tenant-id",
        "preferred_username": "jwt-user@example.com",
        "name": "JWT User",
        "scp": "access_as_user",
        "aud": "api-client-id",
        "iss": "https://login.microsoftonline.com/tenant-id/v2.0",
        "iat": now,
        "exp": now + timedelta(minutes=5),
    }
    token = jwt.encode(claims, private_key, algorithm="RS256", headers={"kid": "test-key"})
    return token, public_key


@pytest.mark.django_db
@override_settings(ENTRA_TENANT_ID="tenant-id", ENTRA_ALLOWED_AUDIENCES=["api-client-id"])
def test_valid_entra_token_is_accepted(signed_token):
    token, public_key = signed_token

    with patch(
        "apps.identity.authentication.jwt.PyJWKClient",
        return_value=SimpleNamespace(get_signing_key_from_jwt=lambda _: SimpleNamespace(key=public_key)),
    ):
        request = APIRequestFactory().get("/api/v1/me", HTTP_AUTHORIZATION=f"Bearer {token}")
        user, claims = EntraBearerAuthentication().authenticate(Request(request))

    assert user.email == "jwt-user@example.com"
    assert claims["oid"] == "jwt-user-object-id"


@pytest.mark.django_db
@override_settings(ENTRA_TENANT_ID="tenant-id", ENTRA_ALLOWED_AUDIENCES=["api-client-id"])
def test_entra_token_without_required_scope_is_rejected(signed_token):
    token, public_key = signed_token
    decoded = jwt.decode(token, options={"verify_signature": False})
    decoded["scp"] = "User.Read"
    private_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
    token = jwt.encode(decoded, private_key, algorithm="RS256", headers={"kid": "test-key"})

    with patch(
        "apps.identity.authentication.jwt.PyJWKClient",
        return_value=SimpleNamespace(get_signing_key_from_jwt=lambda _: SimpleNamespace(key=private_key.public_key())),
    ):
        request = APIRequestFactory().get("/api/v1/me", HTTP_AUTHORIZATION=f"Bearer {token}")
        with pytest.raises(AuthenticationFailed, match="required API scope"):
            EntraBearerAuthentication().authenticate(Request(request))
