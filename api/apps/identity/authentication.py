from __future__ import annotations

import time
from typing import Any

import jwt
from django.conf import settings
from rest_framework.authentication import BaseAuthentication, get_authorization_header
from rest_framework.exceptions import AuthenticationFailed

from .models import PlatformUser


class EntraBearerAuthentication(BaseAuthentication):
    """Validate Entra access tokens and materialize the caller as a PlatformUser."""

    def authenticate(self, request):
        parts = get_authorization_header(request).split()
        if not parts:
            return None
        if len(parts) != 2 or parts[0].lower() != b"bearer":
            raise AuthenticationFailed("Authorization must use a bearer token")

        claims = self._decode(parts[1].decode("utf-8"))
        object_id = claims.get("oid")
        if not object_id:
            raise AuthenticationFailed("Token is missing the Entra object id")

        user, _ = PlatformUser.objects.update_or_create(
            entra_object_id=object_id,
            defaults={
                "tenant_id": claims["tid"],
                "email": claims.get("preferred_username", claims.get("email", "")),
                "display_name": claims.get("name", ""),
            },
        )
        return user, claims

    def authenticate_header(self, request):
        return "Bearer"

    def _decode(self, raw_token: str) -> dict[str, Any]:
        tenant_id = settings.ENTRA_TENANT_ID
        audiences = settings.ENTRA_ALLOWED_AUDIENCES
        if not tenant_id or not audiences:
            raise AuthenticationFailed("Entra authentication is not configured")

        issuer = f"https://login.microsoftonline.com/{tenant_id}/v2.0"
        jwks_client = jwt.PyJWKClient(
            f"https://login.microsoftonline.com/{tenant_id}/discovery/v2.0/keys"
        )
        try:
            signing_key = jwks_client.get_signing_key_from_jwt(raw_token)
            claims = jwt.decode(
                raw_token,
                signing_key.key,
                algorithms=["RS256"],
                audience=audiences,
                issuer=issuer,
                options={"require": ["exp", "iat", "iss", "aud", "tid"]},
            )
        except Exception as exc:
            # Authentication errors should never leak token/JWKS details to callers.
            raise AuthenticationFailed("Invalid Entra access token") from exc

        if claims.get("tid") != tenant_id:
            raise AuthenticationFailed("Token belongs to an unexpected tenant")
        if claims.get("exp", 0) <= time.time():
            raise AuthenticationFailed("Token has expired")

        scopes = set(str(claims.get("scp", "")).split())
        if settings.ENTRA_REQUIRED_SCOPE not in scopes:
            raise AuthenticationFailed("Token does not grant the required API scope")
        return claims
