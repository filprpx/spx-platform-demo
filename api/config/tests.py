from django.conf import settings

from config.celery import app


def test_celery_uses_configured_broker_and_provisioning_queue():
    assert app.conf.broker_url == settings.CELERY_BROKER_URL
    assert app.conf.task_default_queue == "provisioning"


def test_celery_accepts_and_serializes_json_only():
    assert app.conf.task_serializer == "json"
    assert app.conf.result_serializer == "json"
    assert app.conf.accept_content == ["json"]
