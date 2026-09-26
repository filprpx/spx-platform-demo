from rest_framework import generics

from .models import Application
from .serializers import ApplicationSerializer


class ApplicationListCreateView(generics.ListCreateAPIView):
    queryset = Application.objects.select_related("created_by", "provisioning_request")
    serializer_class = ApplicationSerializer


class ApplicationDetailView(generics.RetrieveAPIView):
    queryset = Application.objects.select_related("created_by", "provisioning_request")
    serializer_class = ApplicationSerializer
    lookup_url_kwarg = "application_id"
