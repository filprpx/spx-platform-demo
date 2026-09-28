from apps.applications.models import Application
from apps.provisioning.models import ProvisioningRequest

from .contracts import ProvisioningJob


JOB_SCHEMA_VERSION = 2
JOB_TYPE = "provision_application"


def build_provisioning_job(application: Application, request: ProvisioningRequest) -> dict:
    return ProvisioningJob(
        schema_version=JOB_SCHEMA_VERSION,
        job_type=JOB_TYPE,
        job_id=request.id,
        provisioning_request_id=request.id,
        application={
            "id": application.id,
            "name": application.name,
            "owning_team": application.owning_team,
            "compute_size": application.compute_size,
            "container_port": application.container_port,
            "ingress": application.ingress,
            "min_replicas": application.min_replicas,
            "max_replicas": application.max_replicas,
        },
        requested_by={
            "platform_user_id": request.requested_by.id,
            "entra_object_id": request.requested_by.entra_object_id,
        },
    ).model_dump(mode="json")
