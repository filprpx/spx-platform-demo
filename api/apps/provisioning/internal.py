import hmac

from django.conf import settings
from django.db import transaction
from django.utils import timezone
from rest_framework import serializers, status
from rest_framework.response import Response
from rest_framework.views import APIView

from .models import ProvisioningRequest


class ExecutionEventSerializer(serializers.Serializer):
    schema_version = serializers.IntegerField()
    job_id = serializers.UUIDField()
    provisioning_request_id = serializers.UUIDField()
    status = serializers.ChoiceField(
        choices=[
            ProvisioningRequest.Status.RUNNING,
            ProvisioningRequest.Status.READY_FOR_EXECUTION,
            ProvisioningRequest.Status.WAITING_FOR_APPROVAL,
            ProvisioningRequest.Status.APPLYING,
            ProvisioningRequest.Status.SUCCEEDED,
            ProvisioningRequest.Status.WORKER_FAILED,
            ProvisioningRequest.Status.EXECUTION_FAILED,
        ]
    )
    message = serializers.CharField(required=False, allow_blank=True, max_length=2000)

    def validate_schema_version(self, value):
        if value != 2:
            raise serializers.ValidationError("unsupported schema_version")
        return value


class ExecutionEventView(APIView):
    authentication_classes = []
    permission_classes = []

    def post(self, request):
        configured_token = settings.WORKER_CALLBACK_TOKEN
        supplied_token = request.headers.get("X-Worker-Token", "")
        if not configured_token or not hmac.compare_digest(supplied_token, configured_token):
            return Response({"detail": "Invalid worker token"}, status=status.HTTP_401_UNAUTHORIZED)

        serializer = ExecutionEventSerializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        event = serializer.validated_data
        if str(event["job_id"]) != str(event["provisioning_request_id"]):
            return Response({"detail": "job_id does not match provisioning_request_id"}, status=status.HTTP_409_CONFLICT)

        with transaction.atomic():
            try:
                provisioning_request = ProvisioningRequest.objects.select_for_update().get(
                    pk=event["provisioning_request_id"]
                )
            except ProvisioningRequest.DoesNotExist:
                return Response({"detail": "Provisioning request not found"}, status=status.HTTP_404_NOT_FOUND)

            current = provisioning_request.status
            requested = event["status"]
            if current == requested:
                return Response({"status": current})
            allowed = {
                ProvisioningRequest.Status.QUEUED: {
                    ProvisioningRequest.Status.RUNNING,
                    ProvisioningRequest.Status.WORKER_FAILED,
                },
                ProvisioningRequest.Status.RUNNING: {
                    ProvisioningRequest.Status.READY_FOR_EXECUTION,
                    ProvisioningRequest.Status.WORKER_FAILED,
                },
                ProvisioningRequest.Status.READY_FOR_EXECUTION: {
                    ProvisioningRequest.Status.WAITING_FOR_APPROVAL,
                    ProvisioningRequest.Status.WORKER_FAILED,
                },
                ProvisioningRequest.Status.WAITING_FOR_APPROVAL: {
                    ProvisioningRequest.Status.APPLYING,
                    ProvisioningRequest.Status.EXECUTION_FAILED,
                },
                ProvisioningRequest.Status.APPLYING: {
                    ProvisioningRequest.Status.SUCCEEDED,
                    ProvisioningRequest.Status.EXECUTION_FAILED,
                },
            }
            if requested not in allowed.get(current, set()):
                return Response({"detail": f"Invalid status transition: {current} -> {requested}"}, status=status.HTTP_409_CONFLICT)

            provisioning_request.status = requested
            if requested == ProvisioningRequest.Status.RUNNING:
                provisioning_request.started_at = provisioning_request.started_at or timezone.now()
            if requested in {
                ProvisioningRequest.Status.SUCCEEDED,
                ProvisioningRequest.Status.WORKER_FAILED,
                ProvisioningRequest.Status.EXECUTION_FAILED,
            }:
                provisioning_request.completed_at = timezone.now()
            if requested in {ProvisioningRequest.Status.WORKER_FAILED, ProvisioningRequest.Status.EXECUTION_FAILED}:
                provisioning_request.error = event.get("message", "Worker failed")
            provisioning_request.save()
        return Response({"status": requested})
