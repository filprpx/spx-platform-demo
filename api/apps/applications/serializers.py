import re

from django.db import IntegrityError
from rest_framework import serializers

from apps.identity.serializers import PlatformUserSerializer
from apps.provisioning.models import ProvisioningRequest
from apps.provisioning.services import create_application_with_provisioning_request
from .models import Application
from apps.provisioning.serializers import ProvisioningRequestSerializer


class ApplicationSerializer(serializers.ModelSerializer):
    created_by = PlatformUserSerializer(read_only=True)
    provisioning_request = serializers.SerializerMethodField()

    class Meta:
        model = Application
        fields = [
            "id",
            "name",
            "owning_team",
            "compute_size",
            "container_port",
            "ingress",
            "min_replicas",
            "max_replicas",
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

    def validate_compute_size(self, value):
        if value not in {Application.ComputeSize.SMALL, Application.ComputeSize.MEDIUM}:
            raise serializers.ValidationError("Compute size must be one of: small, medium.")
        return value

    def validate_container_port(self, value):
        if not 1 <= value <= 65535:
            raise serializers.ValidationError("Container port must be between 1 and 65535.")
        return value

    def validate_ingress(self, value):
        if value not in {Application.Ingress.EXTERNAL, Application.Ingress.INTERNAL}:
            raise serializers.ValidationError("Ingress must be one of: external, internal.")
        return value

    def validate(self, attrs):
        min_replicas = attrs.get("min_replicas", 0)
        max_replicas = attrs.get("max_replicas", 1)
        if max_replicas < 1 or max_replicas > 5:
            raise serializers.ValidationError({"max_replicas": "Maximum replicas must be between 1 and 5."})
        if min_replicas > max_replicas:
            raise serializers.ValidationError({"min_replicas": "Minimum replicas cannot exceed maximum replicas."})
        return attrs

    def get_provisioning_request(self, obj):
        request = ProvisioningRequest.objects.filter(application=obj).first()
        return ProvisioningRequestSerializer(request).data if request else None

    def create(self, validated_data):
        user = self.context["request"].user
        try:
            return create_application_with_provisioning_request(validated_data=validated_data, user=user)
        except IntegrityError as exc:
            raise serializers.ValidationError({"name": "An application with this name already exists."}) from exc
