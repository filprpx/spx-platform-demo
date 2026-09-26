package platform

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestCreateApplicationSendsBearerTokenAndPayload(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/applications" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing bearer token")
		}
		return &http.Response{StatusCode: http.StatusCreated, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"id":"app-id","name":"payments-api","type":"api","runtime":"go","owning_team":"finance","created_by":{"email":"developer@example.com"},"provisioning_request":{"id":"request-id","status":"PENDING"}}`)), Request: r}, nil
	})
	client := NewClientWithHTTPClient("http://platform.test", oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}), &http.Client{Transport: transport})
	application, err := client.CreateApplication(context.Background(), CreateApplicationRequest{Name: "payments-api", Type: "api", Runtime: "go", OwningTeam: "finance"})
	if err != nil {
		t.Fatal(err)
	}
	if application.Name != "payments-api" || application.ProvisioningRequest.Status != "PENDING" {
		t.Fatalf("unexpected response: %+v", application)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
