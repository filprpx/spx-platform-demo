package platform

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
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
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request payload: %v", err)
		}
		expected := map[string]any{
			"name": "payments-api", "owning_team": "finance", "compute_size": "small",
			"container_port": float64(8080), "ingress": "external", "min_replicas": float64(0), "max_replicas": float64(1),
		}
		if !reflect.DeepEqual(payload, expected) {
			t.Fatalf("unexpected request payload: %#v", payload)
		}
		return &http.Response{StatusCode: http.StatusCreated, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"id":"app-id","name":"payments-api","owning_team":"finance","compute_size":"small","container_port":8080,"ingress":"external","min_replicas":0,"max_replicas":1,"created_by":{"email":"developer@example.com"},"provisioning_request":{"id":"request-id","status":"PENDING"}}`)), Request: r}, nil
	})
	client := NewClientWithHTTPClient("http://platform.test", oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}), &http.Client{Transport: transport})
	application, err := client.CreateApplication(context.Background(), CreateApplicationRequest{Name: "payments-api", OwningTeam: "finance", ComputeSize: "small", ContainerPort: 8080, Ingress: "external", MinReplicas: 0, MaxReplicas: 1})
	if err != nil {
		t.Fatal(err)
	}
	if application.Name != "payments-api" || application.ProvisioningRequest.Status != "PENDING" {
		t.Fatalf("unexpected response: %+v", application)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
