import uuid

import django.db.models.deletion
from django.db import migrations, models


class Migration(migrations.Migration):
    initial = True
    dependencies = [("identity", "0001_initial")]
    operations = [
        migrations.CreateModel(
            name="Application",
            fields=[
                ("id", models.UUIDField(default=uuid.uuid4, editable=False, primary_key=True, serialize=False)),
                ("name", models.SlugField(max_length=63, unique=True)),
                ("type", models.CharField(max_length=32)),
                ("runtime", models.CharField(max_length=32)),
                ("owning_team", models.CharField(max_length=128)),
                ("created_at", models.DateTimeField(auto_now_add=True)),
                (
                    "created_by",
                    models.ForeignKey(
                        on_delete=django.db.models.deletion.PROTECT,
                        related_name="applications",
                        to="identity.platformuser",
                    ),
                ),
            ],
            options={"ordering": ["name"]},
        ),
    ]
