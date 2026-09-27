package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestCallbackServerReturnsAuthorizationCode(t *testing.T) {
	server, err := NewCallbackServer("http://127.0.0.1:0/callback", "expected-state")
	if err != nil {
		t.Fatal(err)
	}
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	defer server.Close(context.Background())

	query := url.Values{"state": {"expected-state"}, "code": {"authorization-code"}}
	response, err := http.Get(fmt.Sprintf("http://%s/callback?%s", server.listener.Addr(), query.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if _, err := io.ReadAll(response.Body); err != nil {
		t.Fatal(err)
	}

	code, err := server.Wait(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if code != "authorization-code" {
		t.Fatalf("expected authorization code, got %q", code)
	}
}

func TestCallbackServerRejectsInvalidState(t *testing.T) {
	server, err := NewCallbackServer("http://127.0.0.1:0/callback", "expected-state")
	if err != nil {
		t.Fatal(err)
	}
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	defer server.Close(context.Background())

	query := url.Values{"state": {"wrong-state"}, "code": {"authorization-code"}}
	response, err := http.Get(fmt.Sprintf("http://%s/callback?%s", server.listener.Addr(), query.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()

	waitContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := server.Wait(waitContext); err == nil || err.Error() != "OAuth state mismatch" {
		t.Fatalf("expected OAuth state mismatch, got %v", err)
	}
}

func TestCallbackServerHonorsContextCancellation(t *testing.T) {
	server, err := NewCallbackServer("http://127.0.0.1:0/callback", "expected-state")
	if err != nil {
		t.Fatal(err)
	}
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	defer server.Close(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := server.Wait(ctx); err != context.Canceled {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
