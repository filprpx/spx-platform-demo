package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
)

type CallbackServer struct {
	server    *http.Server
	listener  net.Listener
	address   string
	result    chan callbackResult
	closeOnce sync.Once
}

type callbackResult struct {
	code string
	err  error
}

func NewCallbackServer(redirectURI, expectedState string) (*CallbackServer, error) {
	parsed, err := url.Parse(redirectURI)
	if err != nil {
		return nil, fmt.Errorf("parse redirect URI: %w", err)
	}
	if parsed.Host == "" {
		return nil, errors.New("redirect URI must include a host")
	}

	callback := &CallbackServer{
		address: parsed.Host,
		result:  make(chan callbackResult, 1),
	}
	callback.server = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callback.handle(w, r, parsed.Path, expectedState)
		}),
	}
	return callback, nil
}

func (s *CallbackServer) Start() error {
	listener, err := net.Listen("tcp", s.serverAddr())
	if err != nil {
		return fmt.Errorf("listen for OAuth callback: %w", err)
	}
	s.listener = listener
	go func() { _ = s.server.Serve(listener) }()
	return nil
}

func (s *CallbackServer) Wait(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case result := <-s.result:
		return result.code, result.err
	}
}

func (s *CallbackServer) Close(ctx context.Context) error {
	var err error
	s.closeOnce.Do(func() {
		if s.server != nil {
			err = s.server.Shutdown(ctx)
		}
	})
	return err
}

func (s *CallbackServer) serverAddr() string {
	return s.address
}

func (s *CallbackServer) handle(w http.ResponseWriter, r *http.Request, expectedPath, expectedState string) {
	if r.URL.Path != expectedPath {
		http.Error(w, "invalid callback path", http.StatusBadRequest)
		return
	}
	if callbackState := r.URL.Query().Get("state"); callbackState != expectedState {
		s.respond(w, callbackResult{err: errors.New("OAuth state mismatch")}, "Login failed. You can close this window.")
		return
	}
	if callbackError := r.URL.Query().Get("error"); callbackError != "" {
		s.respond(w, callbackResult{err: fmt.Errorf("Microsoft Entra login failed: %s", callbackError)}, "Login failed. You can close this window.")
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		s.respond(w, callbackResult{err: errors.New("OAuth callback did not contain an authorization code")}, "Login failed. You can close this window.")
		return
	}
	s.respond(w, callbackResult{code: code}, "Login complete. You can close this window.")
}

func (s *CallbackServer) respond(w http.ResponseWriter, result callbackResult, message string) {
	select {
	case s.result <- result:
	default:
	}
	_, _ = io.WriteString(w, message)
}
