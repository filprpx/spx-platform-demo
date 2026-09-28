from django.db import migrations, models


class Migration(migrations.Migration):
    dependencies = [("applications", "0001_initial")]

    operations = [
        migrations.RemoveField(model_name="application", name="type"),
        migrations.RemoveField(model_name="application", name="runtime"),
        migrations.AddField(
            model_name="application",
            name="compute_size",
            field=models.CharField(
                choices=[("small", "Small"), ("medium", "Medium")],
                default="small",
                max_length=16,
            ),
        ),
        migrations.AddField(
            model_name="application",
            name="container_port",
            field=models.PositiveIntegerField(default=8080),
        ),
        migrations.AddField(
            model_name="application",
            name="ingress",
            field=models.CharField(
                choices=[("external", "External"), ("internal", "Internal")],
                default="external",
                max_length=16,
            ),
        ),
        migrations.AddField(
            model_name="application",
            name="min_replicas",
            field=models.PositiveIntegerField(default=0),
        ),
        migrations.AddField(
            model_name="application",
            name="max_replicas",
            field=models.PositiveIntegerField(default=1),
        ),
    ]
