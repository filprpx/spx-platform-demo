from typing import Literal
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field


class ProvisioningContract(BaseModel):
    model_config = ConfigDict(extra="forbid", str_strip_whitespace=True)


class ApplicationSnapshot(ProvisioningContract):
    id: UUID
    name: str = Field(min_length=1)
    owning_team: str = Field(min_length=1)
    compute_size: Literal["small", "medium"]
    container_port: int = Field(ge=1, le=65535)
    ingress: Literal["external", "internal"]
    min_replicas: int = Field(ge=0, le=5)
    max_replicas: int = Field(ge=1, le=5)

    def model_post_init(self, __context):
        if self.min_replicas > self.max_replicas:
            raise ValueError("min_replicas cannot exceed max_replicas")


class RequestedBy(ProvisioningContract):
    platform_user_id: UUID
    entra_object_id: str = Field(min_length=1)


class ProvisioningJob(ProvisioningContract):
    schema_version: Literal[2]
    job_type: Literal["provision_application"]
    job_id: UUID
    provisioning_request_id: UUID
    application: ApplicationSnapshot
    requested_by: RequestedBy

    def model_post_init(self, __context):
        if self.job_id != self.provisioning_request_id:
            raise ValueError("job_id and provisioning_request_id must match")


class JobIdentifiers(ProvisioningContract):
    job_id: UUID
    provisioning_request_id: UUID


def extract_job_identifiers(payload: object) -> tuple[str, str] | None:
    if not isinstance(payload, dict):
        return None
    try:
        identifiers = JobIdentifiers.model_validate(
            {
                "job_id": payload.get("job_id"),
                "provisioning_request_id": payload.get("provisioning_request_id"),
            }
        )
    except ValueError:
        return None
    if identifiers.job_id != identifiers.provisioning_request_id:
        return None
    return str(identifiers.job_id), str(identifiers.provisioning_request_id)
