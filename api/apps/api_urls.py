from django.urls import path

from apps.applications.views import ApplicationDetailView, ApplicationListCreateView
from apps.identity.views import MeView
from apps.provisioning.views import ProvisioningRequestDetailView


urlpatterns = [
    path("me", MeView.as_view(), name="me"),
    path("applications", ApplicationListCreateView.as_view(), name="applications"),
    path("applications/<uuid:application_id>", ApplicationDetailView.as_view(), name="application-detail"),
    path("provisioning/<uuid:request_id>", ProvisioningRequestDetailView.as_view(), name="provisioning-detail"),
]
