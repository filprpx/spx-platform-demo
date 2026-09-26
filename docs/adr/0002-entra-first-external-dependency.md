# Use Microsoft Entra for Milestone 1 authentication

Milestone 1 uses Microsoft Entra for real developer authentication. The CLI performs a browser-based authorization-code flow with PKCE, and the Django API validates Entra bearer tokens before mapping the identity to a local `PlatformUser`.

Terraform bootstraps the Entra application registrations, API scope, and CLI permissions. This bootstrap configuration is separate from application workload provisioning.

Persistence and runtime remain local with SQLite, Docker Compose, and Django Admin. The normal demo flow uses real Entra authentication; a development authentication mode is not part of the milestone. This proves identity propagation without requiring Azure workload, pipeline, queue, or production database infrastructure.
