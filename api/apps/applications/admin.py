from django.contrib import admin

from .models import Application


@admin.register(Application)
class ApplicationAdmin(admin.ModelAdmin):
    list_display = ["name", "compute_size", "container_port", "ingress", "owning_team", "created_by", "created_at"]
    list_filter = ["compute_size", "ingress"]
    search_fields = ["name", "owning_team", "created_by__email"]
