# Internal Developer Platform

This context captures the language of the authenticated control plane. It records developer intent and identity; it does not describe Azure resources or provisioning implementation.

## Identity and ownership

**Developer**:
The human operating the CLI.
_Avoid_: Operator, requester (when referring to the human generally)

**Platform User**:
The local representation of an authenticated Microsoft Entra identity.
_Avoid_: User account, developer account

**Owning Team**:
The stable team identifier accountable for an Application.
_Avoid_: Owner, owning user

## Platform intent

**Application**:
A registered workload intent with an owning team and creator; it does not imply that an Azure resource exists.
_Avoid_: Service, deployment, Azure application

**Provisioning Request**:
A recorded request to eventually materialize an Application through an execution plane.
_Avoid_: Job, deployment, Azure operation

**Pending**:
The initial Provisioning Request state meaning the platform accepted the intent but execution has not started.
_Avoid_: Queued, processing, ready
