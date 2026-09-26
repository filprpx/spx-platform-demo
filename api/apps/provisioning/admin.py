from django.contrib import admin

from .models import ProvisioningRequest


@admin.register(ProvisioningRequest)
class ProvisioningRequestAdmin(admin.ModelAdmin):
    list_display = ["application", "status", "requested_by", "created_at", "completed_at"]
    list_filter = ["status"]
    search_fields = ["application__name", "requested_by__email"]
