from rest_framework import serializers

from .models import PlatformUser


class PlatformUserSerializer(serializers.ModelSerializer):
    class Meta:
        model = PlatformUser
        fields = ["id", "entra_object_id", "tenant_id", "email", "display_name", "last_seen_at"]
