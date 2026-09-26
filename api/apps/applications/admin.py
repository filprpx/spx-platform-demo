from django.contrib import admin

from .models import Application


@admin.register(Application)
class ApplicationAdmin(admin.ModelAdmin):
    list_display = ["name", "type", "runtime", "owning_team", "created_by", "created_at"]
    list_filter = ["type", "runtime"]
    search_fields = ["name", "owning_team", "created_by__email"]
