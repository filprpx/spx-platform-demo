package platform

type Application struct {
	ID                  string               `json:"id"`
	Name                string               `json:"name"`
	Type                string               `json:"type"`
	Runtime             string               `json:"runtime"`
	OwningTeam          string               `json:"owning_team"`
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
	Name       string `json:"name"`
	Type       string `json:"type"`
	Runtime    string `json:"runtime"`
	OwningTeam string `json:"owning_team"`
}
