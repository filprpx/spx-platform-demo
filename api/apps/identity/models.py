import uuid

from django.db import models


class PlatformUser(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    entra_object_id = models.CharField(max_length=128, unique=True)
    tenant_id = models.CharField(max_length=128)
    email = models.EmailField(blank=True)
    display_name = models.CharField(max_length=255, blank=True)
    last_seen_at = models.DateTimeField(auto_now=True)
    created_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ["email", "entra_object_id"]

    def __str__(self) -> str:
        return self.email or self.display_name or self.entra_object_id

    @property
    def is_authenticated(self) -> bool:
        return True

    @property
    def is_anonymous(self) -> bool:
        return False
