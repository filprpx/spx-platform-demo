from django.contrib import admin
from django.urls import include, path

from apps.provisioning.internal import ExecutionEventView


urlpatterns = [
    path("admin/", admin.site.urls),
    path("api/v1/", include("apps.api_urls")),
    path("internal/execution-events", ExecutionEventView.as_view(), name="execution-events"),
]
