package auth

import "testing"

func TestBrowserCommandForOS(t *testing.T) {
	tests := []struct {
		name    string
		os      string
		command string
		args    []string
	}{
		{name: "linux", os: "linux", command: "xdg-open", args: []string{"https://example.test"}},
		{name: "macos", os: "darwin", command: "open", args: []string{"https://example.test"}},
		{name: "windows", os: "windows", command: "rundll32", args: []string{"url.dll,FileProtocolHandler", "https://example.test"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command, args, err := browserCommandForOS(test.os, "https://example.test")
			if err != nil {
				t.Fatal(err)
			}
			if command != test.command {
				t.Fatalf("expected command %q, got %q", test.command, command)
			}
			if len(args) != len(test.args) {
				t.Fatalf("expected args %v, got %v", test.args, args)
			}
			for i := range args {
				if args[i] != test.args[i] {
					t.Fatalf("expected args %v, got %v", test.args, args)
				}
			}
		})
	}
}

func TestBrowserCommandForUnsupportedOS(t *testing.T) {
	if _, _, err := browserCommandForOS("plan9", "https://example.test"); err == nil {
		t.Fatal("expected unsupported OS error")
	}
}

func TestBrowserLauncherUsesInjectedProcessStarter(t *testing.T) {
	var command string
	var args []string
	launcher := BrowserLauncher{
		OS: "linux",
		Start: func(gotCommand string, gotArgs ...string) error {
			command = gotCommand
			args = gotArgs
			return nil
		},
	}

	if err := launcher.Open("https://example.test"); err != nil {
		t.Fatal(err)
	}
	if command != "xdg-open" {
		t.Fatalf("expected xdg-open, got %q", command)
	}
	if len(args) != 1 || args[0] != "https://example.test" {
		t.Fatalf("unexpected args: %v", args)
	}
}
