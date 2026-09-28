import uuid

from django.db import models

from apps.identity.models import PlatformUser


class ProvisioningRequest(models.Model):
    class Status(models.TextChoices):
        PENDING = "PENDING", "Pending"
        QUEUED = "QUEUED", "Queued"
        QUEUE_FAILED = "QUEUE_FAILED", "Queue failed"
        RUNNING = "RUNNING", "Running"
        READY_FOR_EXECUTION = "READY_FOR_EXECUTION", "Ready for execution"
        WAITING_FOR_APPROVAL = "WAITING_FOR_APPROVAL", "Waiting for approval"
        APPLYING = "APPLYING", "Applying"
        SUCCEEDED = "SUCCEEDED", "Succeeded"
        WORKER_FAILED = "WORKER_FAILED", "Worker failed"
        EXECUTION_FAILED = "EXECUTION_FAILED", "Execution failed"

    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    application = models.OneToOneField(
        "applications.Application",
        on_delete=models.CASCADE,
        related_name="provisioning_request",
    )
    status = models.CharField(max_length=32, choices=Status.choices, default=Status.PENDING)
    requested_by = models.ForeignKey(PlatformUser, on_delete=models.PROTECT, related_name="provisioning_requests")
    created_at = models.DateTimeField(auto_now_add=True)
    started_at = models.DateTimeField(null=True, blank=True)
    completed_at = models.DateTimeField(null=True, blank=True)
    error = models.TextField(blank=True)

    class Meta:
        ordering = ["-created_at"]

    def __str__(self) -> str:
        return f"{self.application.name} ({self.status})"
