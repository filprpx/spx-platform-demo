from django.db import migrations, models


class Migration(migrations.Migration):
    dependencies = [
        ("provisioning", "0002_provisioning_statuses"),
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
                    ("RUNNING", "Running"),
                    ("READY_FOR_EXECUTION", "Ready for execution"),
                    ("WORKER_FAILED", "Worker failed"),
                ],
                default="PENDING",
                max_length=32,
            ),
        ),
    ]
