package platform

import (
	"os"
	"testing"
)

// TestApplyEnvDefaults_SetsUnsetKeys verifies that ApplyEnvDefaults populates
// env vars that the user hasn't set. Runs on both Linux and Windows — the
// exact values differ per OS but the behavioral guarantee is identical.
func TestApplyEnvDefaults_SetsUnsetKeys(t *testing.T) {
	keys := []string{
		"CONTAINERD_SOCK",
		"CONTAINERD_CONTENT_PATH",
		"CONTAINERD_REGISTRY_CONFIG_PATH",
	}
	for _, k := range keys {
		t.Setenv(k, "") // Setenv with empty then Unsetenv to reset cleanly
		os.Unsetenv(k)
	}
	ApplyEnvDefaults()
	d := Defaults()
	if got := os.Getenv("CONTAINERD_SOCK"); got != d.ContainerdSock {
		t.Errorf("CONTAINERD_SOCK = %q, want %q", got, d.ContainerdSock)
	}
	if got := os.Getenv("CONTAINERD_CONTENT_PATH"); got != d.ContainerdContentPath {
		t.Errorf("CONTAINERD_CONTENT_PATH = %q, want %q", got, d.ContainerdContentPath)
	}
	if got := os.Getenv("CONTAINERD_REGISTRY_CONFIG_PATH"); got != d.ContainerdRegistryConfigPath {
		t.Errorf("CONTAINERD_REGISTRY_CONFIG_PATH = %q, want %q", got, d.ContainerdRegistryConfigPath)
	}
}

// TestApplyEnvDefaults_RespectsUserOverride ensures user-set env vars win over
// our platform defaults. This is the critical correctness property that lets
// a Linux operator run Spegel pointing at a bespoke socket path.
func TestApplyEnvDefaults_RespectsUserOverride(t *testing.T) {
	const custom = "/tmp/my-custom-containerd.sock"
	t.Setenv("CONTAINERD_SOCK", custom)
	ApplyEnvDefaults()
	if got := os.Getenv("CONTAINERD_SOCK"); got != custom {
		t.Errorf("CONTAINERD_SOCK = %q, want %q (user override must win)", got, custom)
	}
}
