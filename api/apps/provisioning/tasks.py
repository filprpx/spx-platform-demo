from celery import shared_task
from pathlib import Path

from .contracts import ProvisioningJob, extract_job_identifiers
from .desired_state import write_desired_state
from .worker_client import CallbackDeliveryError, post_execution_event


@shared_task(
    bind=True,
    name="provisioning.execute",
    acks_late=True,
    reject_on_worker_lost=True,
    autoretry_for=(CallbackDeliveryError,),
    retry_backoff=True,
    retry_backoff_max=30,
    retry_jitter=False,
    retry_kwargs={"max_retries": 3},
)
def execute_provisioning_job(self, payload):
    try:
        job = ProvisioningJob.model_validate(payload)
    except (TypeError, ValueError) as exc:
        identifiers = extract_job_identifiers(payload)
        if identifiers is not None:
            job_id, request_id = identifiers
            post_execution_event(
                job_id=job_id,
                provisioning_request_id=request_id,
                status="WORKER_FAILED",
                message=str(exc),
            )
        return {"status": "WORKER_FAILED", "message": str(exc)}

    job_id = str(job.job_id)
    request_id = str(job.provisioning_request_id)
    post_execution_event(
        job_id=job_id,
        provisioning_request_id=request_id,
        status="RUNNING",
        message="Worker accepted provisioning job",
    )
    post_execution_event(
        job_id=job_id,
        provisioning_request_id=request_id,
        status="READY_FOR_EXECUTION",
        message="Worker accepted provisioning job",
    )
    try:
        write_desired_state(job, Path(__file__).resolve().parents[3])
    except Exception as exc:
        post_execution_event(
            job_id=job_id,
            provisioning_request_id=request_id,
            status="WORKER_FAILED",
            message=str(exc),
        )
        return {"status": "WORKER_FAILED", "message": str(exc)}
    post_execution_event(
        job_id=job_id,
        provisioning_request_id=request_id,
        status="WAITING_FOR_APPROVAL",
        message="Terraform desired state is ready; iac-pipeline represents approved PR execution",
    )
    return {"status": "WAITING_FOR_APPROVAL", "job_id": job_id}
