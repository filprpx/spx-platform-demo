import uuid

from django.db import migrations, models


class Migration(migrations.Migration):
    initial = True
    dependencies = []
    operations = [
        migrations.CreateModel(
            name="PlatformUser",
            fields=[
                ("id", models.UUIDField(default=uuid.uuid4, editable=False, primary_key=True, serialize=False)),
                ("entra_object_id", models.CharField(max_length=128, unique=True)),
                ("tenant_id", models.CharField(max_length=128)),
                ("email", models.EmailField(blank=True, max_length=254)),
                ("display_name", models.CharField(blank=True, max_length=255)),
                ("last_seen_at", models.DateTimeField(auto_now=True)),
                ("created_at", models.DateTimeField(auto_now_add=True)),
            ],
            options={"ordering": ["email", "entra_object_id"]},
        ),
    ]
