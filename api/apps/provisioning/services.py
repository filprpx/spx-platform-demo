from django.db import transaction

from apps.applications.models import Application
from apps.identity.models import PlatformUser

from .job_contract import build_provisioning_job
from .models import ProvisioningRequest
from .publisher import ProvisioningQueueError, publish_provisioning_job


def create_application_with_provisioning_request(
    *,
    validated_data: dict,
    user: PlatformUser,
    publisher=None,
) -> Application:
    if publisher is None:
        publisher = publish_provisioning_job
    with transaction.atomic():
        application = Application.objects.create(created_by=user, **validated_data)
        request = ProvisioningRequest.objects.create(
            application=application,
            requested_by=user,
            status=ProvisioningRequest.Status.PENDING,
        )
        payload = build_provisioning_job(application, request)

        def publish_after_commit():
            try:
                publisher(payload)
            except ProvisioningQueueError as exc:
                ProvisioningRequest.objects.filter(pk=request.pk).update(
                    status=ProvisioningRequest.Status.QUEUE_FAILED,
                    error=str(exc.detail),
                )
                raise
            else:
                ProvisioningRequest.objects.filter(pk=request.pk).update(
                    status=ProvisioningRequest.Status.QUEUED,
                )

        transaction.on_commit(publish_after_commit)
    return application
