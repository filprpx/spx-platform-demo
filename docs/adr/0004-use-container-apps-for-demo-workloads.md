# Use Azure Container Apps for demo workloads

The first workload target is Azure Container Apps rather than an Azure Virtual Machine.

Developers describe infrastructure intent using platform-level settings:

- compute size (`small` or `medium`);
- container port;
- public or internal ingress;
- minimum replicas;
- maximum replicas.

The API does not expose an Azure-specific template selector, and it does not ask developers for Git, Dockerfile, image, or credential information. The worker maps the validated intent to Terraform for a Container App, its Container Apps Environment, Log Analytics support, managed identity, and registry pull permission.

The Container App is created with a small public placeholder image because Azure requires an image reference when creating the resource. It defaults to the Consumption plan with scale-to-zero behavior. The application image is supplied later by the separate local deployment command.

This gives the demo a real managed container workload while avoiding VM operating-system management, SSH, Docker installation, process supervision, public-IP configuration, and always-on compute costs.

VM support remains deferred. A future workload target may be added if the platform needs to demonstrate host-level infrastructure control.
