import uuid

from django.db import models

from apps.identity.models import PlatformUser


class Application(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    name = models.SlugField(max_length=63, unique=True)
    type = models.CharField(max_length=32)
    runtime = models.CharField(max_length=32)
    owning_team = models.CharField(max_length=128)
    created_by = models.ForeignKey(PlatformUser, on_delete=models.PROTECT, related_name="applications")
    created_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ["name"]

    def __str__(self) -> str:
        return self.name
