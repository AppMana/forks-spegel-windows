//go:build linux

package platform

func defaults() Paths {
	return Paths{
		ContainerdSock:               "/run/containerd/containerd.sock",
		ContainerdContentPath:        "/var/lib/containerd/io.containerd.content.v1.content",
		ContainerdRegistryConfigPath: "/etc/containerd/certs.d",
		PersistenceHostPath:          "/var/lib/spegel",
		BasicAuthSecretsDir:          "/etc/secrets/basic-auth",
	}
}
