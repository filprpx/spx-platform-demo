import re

from django.db import IntegrityError, transaction
from rest_framework import serializers

from apps.identity.serializers import PlatformUserSerializer
from .models import Application
from apps.provisioning.models import ProvisioningRequest
from apps.provisioning.serializers import ProvisioningRequestSerializer


class ApplicationSerializer(serializers.ModelSerializer):
    created_by = PlatformUserSerializer(read_only=True)
    provisioning_request = serializers.SerializerMethodField()

    class Meta:
        model = Application
        fields = [
            "id",
            "name",
            "type",
            "runtime",
            "owning_team",
            "created_by",
            "created_at",
            "provisioning_request",
        ]
        read_only_fields = ["id", "created_by", "created_at", "provisioning_request"]

    def validate_name(self, value):
        value = value.strip().lower()
        if not re.fullmatch(r"[a-z][a-z0-9-]{2,62}", value):
            raise serializers.ValidationError(
                "Use 3-63 lowercase characters, starting with a letter; hyphens are allowed."
            )
        return value

    def validate_owning_team(self, value):
        value = value.strip().lower()
        if not re.fullmatch(r"[a-z0-9][a-z0-9-]{1,127}", value):
            raise serializers.ValidationError("Owning team must be a stable lowercase identifier.")
        return value

    def validate_type(self, value):
        if value not in {"api", "worker", "web"}:
            raise serializers.ValidationError("Type must be one of: api, worker, web.")
        return value

    def validate_runtime(self, value):
        value = value.strip().lower()
        if not value:
            raise serializers.ValidationError("Runtime is required.")
        return value

    def get_provisioning_request(self, obj):
        request = getattr(obj, "provisioning_request", None)
        return ProvisioningRequestSerializer(request).data if request else None

    def create(self, validated_data):
        user = self.context["request"].user
        with transaction.atomic():
            try:
                application = Application.objects.create(created_by=user, **validated_data)
            except IntegrityError as exc:
                raise serializers.ValidationError({"name": "An application with this name already exists."}) from exc
            ProvisioningRequest.objects.create(
                application=application,
                requested_by=user,
                status=ProvisioningRequest.Status.PENDING,
            )
        return application
