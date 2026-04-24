//go:build windows

package platform

// Windows defaults derived from observed containerd 2.2.1 layout on Windows
// Server 2022 Datacenter (build 20348.4773). See `probes/appmana-023/findings.md`
// in this fork for the capture.
func defaults() Paths {
	return Paths{
		// containerd/v2 client dials the pipe path as-is; no `npipe://` URI
		// or winio shim is required (verified by cmd/probe-containerd-windows).
		ContainerdSock:               `\\.\pipe\containerd-containerd`,
		ContainerdContentPath:        `C:\ProgramData\containerd\root\io.containerd.content.v1.content`,
		ContainerdRegistryConfigPath: `C:\ProgramData\containerd\certs.d`,
		PersistenceHostPath:          `C:\ProgramData\spegel`,
		BasicAuthSecretsDir:          `C:\spegel\basic-auth`,
	}
}
