import os

from celery import Celery

os.environ.setdefault("DJANGO_SETTINGS_MODULE", "config.settings")

app = Celery("spx")
app.config_from_object("django.conf:settings", namespace="CELERY")
app.autodiscover_tasks(["apps.provisioning"])

# Import the task module explicitly because this app is initialized while Django
# is still loading settings, before autodiscovery can inspect the app registry.
from apps.provisioning import tasks as _provisioning_tasks  # noqa: E402,F401
