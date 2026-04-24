// Package platform exposes the OS-specific default paths that Spegel needs to
// connect to containerd and persist its state. The defaults are baked in at
// build time via `_linux.go` and `_windows.go` files; a single [Defaults]
// accessor returns them.
//
// The defaults apply to flag parsing via environment variables: go-arg honors
// `env:FOO` tags with precedence over struct-tag defaults, so [ApplyEnvDefaults]
// sets each variable iff not already set by the user. The net effect is:
// user-specified env/flag wins → OS-specific default fills the gap →
// Linux-style struct-tag default kicks in as a last-resort fallback (and in
// practice is only reached when running on Linux).
package platform

import "os"

// Paths groups the configurable filesystem and socket addresses that differ
// between Linux and Windows nodes. Every field carries a sensible default for
// the target OS; callers are free to override any of them via CLI flags or
// matching environment variables.
type Paths struct {
	// ContainerdSock is the endpoint address Spegel passes to containerd/v2's
	// Go client. On Linux it's a unix socket path. On Windows it's the raw
	// named pipe name (no `npipe://` URI wrapping — the v2 client handles
	// pipe paths directly, verified against containerd 2.2.1 on Windows
	// Server 2022).
	ContainerdSock string

	// ContainerdContentPath is the on-disk root of containerd's content
	// store. Spegel reads blobs from `<path>/blobs/<algo>/<digest>` when
	// `--containerd-content-path` is non-empty; otherwise it falls back to
	// the gRPC ContentStore API, which works on both OSes.
	ContainerdContentPath string

	// ContainerdRegistryConfigPath is the directory where Spegel writes
	// per-registry `hosts.toml` files that point containerd at the local
	// Spegel registry. Must match containerd's `config_path` setting under
	// the CRI `registry` plugin section.
	ContainerdRegistryConfigPath string

	// PersistenceHostPath is the host directory bind-mounted into the pod
	// where Spegel keeps its libp2p identity across restarts.
	PersistenceHostPath string

	// BasicAuthSecretsDir is the directory containing the `username` and
	// `password` files for basic auth against upstream registries. Mounted
	// into the pod from a Secret volume.
	BasicAuthSecretsDir string
}

// Defaults returns the OS-appropriate [Paths]. On Linux it mirrors the values
// that were historically baked into struct tags; on Windows it reflects the
// containerd 2.x layout observed on a HostProcess-adjacent runtime.
//
// The concrete values live in `platform_linux.go` and `platform_windows.go`,
// selected at build time.
func Defaults() Paths { return defaults() }

// ApplyEnvDefaults populates the Spegel environment variables that the
// [RegistryCmd] struct parses with [Defaults], but only for variables the
// caller hasn't already set. This is called early in `main.main()` so that:
//
//   - user-provided env vars win (no overwrite);
//   - OS-specific defaults apply when running on Windows;
//   - Linux keeps its existing struct-tag defaults (the env sets here are
//     identical to the tag values, so behavior is unchanged).
//
// The mapping below must stay in sync with the `env:` tags on
// [RegistryCmd] and [ConfigurationCmd] in main.go.
func ApplyEnvDefaults() {
	d := Defaults()
	setIfUnset("CONTAINERD_SOCK", d.ContainerdSock)
	setIfUnset("CONTAINERD_CONTENT_PATH", d.ContainerdContentPath)
	setIfUnset("CONTAINERD_REGISTRY_CONFIG_PATH", d.ContainerdRegistryConfigPath)
	// BasicAuthSecretsDir and PersistenceHostPath don't currently have
	// corresponding env flags in Spegel; they're used by the helm chart to
	// mount host paths. Kept here as a single source of truth for the chart
	// template to consume.
}

func setIfUnset(key, value string) {
	if _, ok := os.LookupEnv(key); ok {
		return
	}
	if value == "" {
		return
	}
	_ = os.Setenv(key, value)
}
