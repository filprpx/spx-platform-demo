from rest_framework import serializers

from .models import ProvisioningRequest


class ProvisioningRequestSerializer(serializers.ModelSerializer):
    requested_by_id = serializers.UUIDField(source="requested_by.id", read_only=True)

    class Meta:
        model = ProvisioningRequest
        fields = [
            "id",
            "application",
            "status",
            "requested_by_id",
            "created_at",
            "started_at",
            "completed_at",
            "error",
        ]
