from django.contrib import admin

from .models import PlatformUser


@admin.register(PlatformUser)
class PlatformUserAdmin(admin.ModelAdmin):
    list_display = ["email", "display_name", "entra_object_id", "tenant_id", "last_seen_at"]
    search_fields = ["email", "display_name", "entra_object_id"]
