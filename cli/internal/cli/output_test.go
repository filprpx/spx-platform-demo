package cli

import (
	"bytes"
	"strings"
	"testing"

	"spx/internal/platform"
)

func TestWriteApplicationListIncludesHeadersAndStatus(t *testing.T) {
	var output bytes.Buffer
	writeApplicationList(&output, []platform.Application{{
		Name: "payments-api", Type: "api", Runtime: "go", OwningTeam: "finance",
		ProvisioningRequest: &platform.ProvisioningRequest{Status: "READY"},
	}})

	result := output.String()
	for _, expected := range []string{"NAME", "TYPE", "RUNTIME", "OWNING TEAM", "STATUS", "payments-api", "READY"} {
		if !strings.Contains(result, expected) {
			t.Fatalf("expected %q in output: %q", expected, result)
		}
	}
}

func TestWriteApplicationListDoesNotClipLongStatus(t *testing.T) {
	var output bytes.Buffer
	writeApplicationList(&output, []platform.Application{{
		Name: "payments-api", Type: "api", Runtime: "go", OwningTeam: "finance",
		ProvisioningRequest: &platform.ProvisioningRequest{Status: "READY_FOR_EXECUTION"},
	}})

	if !strings.Contains(output.String(), "READY_FOR_EXECUTION") {
		t.Fatalf("long status was clipped: %q", output.String())
	}
}

func TestWriteApplicationListReportsEmptyState(t *testing.T) {
	var output bytes.Buffer
	writeApplicationList(&output, nil)
	if output.String() != "No applications found.\n" {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestWriteApplicationDetailsHandlesMissingProvisioningRequest(t *testing.T) {
	var output bytes.Buffer
	writeApplicationDetails(&output, platform.Application{Name: "web", ID: "app-1", CreatedBy: platform.PlatformUser{Email: "user@example.test"}})

	result := output.String()
	if !strings.Contains(result, "Provisioning request: -") || !strings.Contains(result, "Provisioning status: not started") {
		t.Fatalf("unexpected output: %q", result)
	}
}

func TestProvisioningSummaryDefaultsEmptyFields(t *testing.T) {
	id, status := provisioningSummary(&platform.ProvisioningRequest{})
	if id != "-" || status != "unknown" {
		t.Fatalf("unexpected summary: %q, %q", id, status)
	}
}
