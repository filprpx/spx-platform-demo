package auth

import "github.com/zalando/go-keyring"

const keyringService = "spx-platform"

type tokenStore interface {
	Get(service, user string) (string, error)
	Set(service, user, password string) error
	Delete(service, user string) error
}

type systemTokenStore struct{}

func (systemTokenStore) Get(service, user string) (string, error) {
	return keyring.Get(service, user)
}

func (systemTokenStore) Set(service, user, password string) error {
	return keyring.Set(service, user, password)
}

func (systemTokenStore) Delete(service, user string) error {
	return keyring.Delete(service, user)
}
