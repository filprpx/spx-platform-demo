import uuid

from django.db import models

from apps.identity.models import PlatformUser


class Application(models.Model):
    class ComputeSize(models.TextChoices):
        SMALL = "small", "Small"
        MEDIUM = "medium", "Medium"

    class Ingress(models.TextChoices):
        EXTERNAL = "external", "External"
        INTERNAL = "internal", "Internal"

    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    name = models.SlugField(max_length=63, unique=True)
    owning_team = models.CharField(max_length=128)
    compute_size = models.CharField(max_length=16, choices=ComputeSize.choices, default=ComputeSize.SMALL)
    container_port = models.PositiveIntegerField(default=8080)
    ingress = models.CharField(max_length=16, choices=Ingress.choices, default=Ingress.EXTERNAL)
    min_replicas = models.PositiveIntegerField(default=0)
    max_replicas = models.PositiveIntegerField(default=1)
    created_by = models.ForeignKey(PlatformUser, on_delete=models.PROTECT, related_name="applications")
    created_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ["name"]

    def __str__(self) -> str:
        return self.name
