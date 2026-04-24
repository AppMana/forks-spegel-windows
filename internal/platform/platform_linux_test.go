//go:build linux

package platform

import "testing"

func TestDefaultsLinux(t *testing.T) {
	d := Defaults()
	want := Paths{
		ContainerdSock:               "/run/containerd/containerd.sock",
		ContainerdContentPath:        "/var/lib/containerd/io.containerd.content.v1.content",
		ContainerdRegistryConfigPath: "/etc/containerd/certs.d",
		PersistenceHostPath:          "/var/lib/spegel",
		BasicAuthSecretsDir:          "/etc/secrets/basic-auth",
	}
	if d != want {
		t.Errorf("Defaults() = %+v, want %+v", d, want)
	}
}
