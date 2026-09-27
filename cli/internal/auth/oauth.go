package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os/exec"
	"runtime"
)

type ProcessStarter func(command string, args ...string) error

type BrowserLauncher struct {
	OS    string
	Start ProcessStarter
}

func NewBrowserLauncher() BrowserLauncher {
	return BrowserLauncher{
		OS: runtime.GOOS,
		Start: func(command string, args ...string) error {
			return exec.Command(command, args...).Start()
		},
	}
}

func randomString(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (b BrowserLauncher) Open(target string) error {
	if b.Start == nil {
		return fmt.Errorf("browser launcher has no process starter")
	}
	command, args, err := browserCommandForOS(b.OS, target)
	if err != nil {
		return err
	}
	return b.Start(command, args...)
}

func browserCommandForOS(osName, target string) (string, []string, error) {
	switch osName {
	case "darwin":
		return "open", []string{target}, nil
	case "linux":
		return "xdg-open", []string{target}, nil
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", target}, nil
	default:
		return "", nil, fmt.Errorf("unsupported operating system: %s", osName)
	}
}
