package auth

import (
	"crypto/rand"
	"encoding/base64"
	"os/exec"
	"runtime"
)

func randomString(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func openBrowser(target string) error {
	command := "xdg-open"
	if runtime.GOOS == "darwin" {
		command = "open"
	}
	if runtime.GOOS == "windows" {
		command = "rundll32"
	}
	args := []string{target}
	if runtime.GOOS == "windows" {
		args = []string{"url.dll,FileProtocolHandler", target}
	}
	return exec.Command(command, args...).Start()
}
