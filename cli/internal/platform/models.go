package platform

type Application struct {
	ID                  string               `json:"id"`
	Name                string               `json:"name"`
	OwningTeam          string               `json:"owning_team"`
	ComputeSize         string               `json:"compute_size"`
	ContainerPort       int                  `json:"container_port"`
	Ingress             string               `json:"ingress"`
	MinReplicas         int                  `json:"min_replicas"`
	MaxReplicas         int                  `json:"max_replicas"`
	CreatedBy           PlatformUser         `json:"created_by"`
	CreatedAt           string               `json:"created_at"`
	ProvisioningRequest *ProvisioningRequest `json:"provisioning_request"`
}

type ProvisioningRequest struct {
	ID            string `json:"id"`
	Application   string `json:"application"`
	Status        string `json:"status"`
	RequestedByID string `json:"requested_by_id"`
	CreatedAt     string `json:"created_at"`
	Error         string `json:"error"`
}

type PlatformUser struct {
	ID            string `json:"id"`
	EntraObjectID string `json:"entra_object_id"`
	TenantID      string `json:"tenant_id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
}

type CreateApplicationRequest struct {
	Name          string `json:"name"`
	OwningTeam    string `json:"owning_team"`
	ComputeSize   string `json:"compute_size"`
	ContainerPort int    `json:"container_port"`
	Ingress       string `json:"ingress"`
	MinReplicas   int    `json:"min_replicas"`
	MaxReplicas   int    `json:"max_replicas"`
}
