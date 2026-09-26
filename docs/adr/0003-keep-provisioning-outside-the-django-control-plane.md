# Keep provisioning execution outside the Django control plane

The Django application is the control plane. It records application intent, owning-team responsibility, creator attribution, provisioning-request records, and status visibility. Creating an application creates one pending provisioning request in the same transaction.

Future execution will be represented through pull requests and independently running automation. Django may initiate or observe that workflow, but it does not execute Terraform or directly provision Azure resources.

This boundary keeps domain state and user-facing orchestration in the API while leaving execution concerns to the later asynchronous, pull-request, pipeline, and Azure provisioning slices. Terraform and Azure execution logic must not be embedded in Django business logic.
