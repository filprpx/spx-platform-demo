from rest_framework import generics

from .models import ProvisioningRequest
from .serializers import ProvisioningRequestSerializer


class ProvisioningRequestDetailView(generics.RetrieveAPIView):
    queryset = ProvisioningRequest.objects.select_related("application", "requested_by")
    serializer_class = ProvisioningRequestSerializer
    lookup_url_kwarg = "request_id"
