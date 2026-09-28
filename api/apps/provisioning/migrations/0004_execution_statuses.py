from django.db import migrations, models


class Migration(migrations.Migration):
    dependencies = [("provisioning", "0003_worker_statuses")]

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
                    ("WAITING_FOR_APPROVAL", "Waiting for approval"),
                    ("APPLYING", "Applying"),
                    ("SUCCEEDED", "Succeeded"),
                    ("WORKER_FAILED", "Worker failed"),
                    ("EXECUTION_FAILED", "Execution failed"),
                ],
                default="PENDING",
                max_length=32,
            ),
        )
    ]
