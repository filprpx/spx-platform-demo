from django.db import migrations, models


class Migration(migrations.Migration):
    dependencies = [
        ("provisioning", "0001_initial"),
    ]

    operations = [
        migrations.AlterField(
            model_name="provisioningrequest",
            name="status",
            field=models.CharField(
                choices=[
                    ("PENDING", "Pending"),
                    ("QUEUED", "Queued"),
                    ("QUEUE_FAILED", "Queue failed"),
                ],
                default="PENDING",
                max_length=32,
            ),
        ),
    ]
