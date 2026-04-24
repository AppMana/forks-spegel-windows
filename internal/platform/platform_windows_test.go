//go:build windows

package platform

import "testing"

func TestDefaultsWindows(t *testing.T) {
	d := Defaults()
	want := Paths{
		ContainerdSock:               `\\.\pipe\containerd-containerd`,
		ContainerdContentPath:        `C:\ProgramData\containerd\root\io.containerd.content.v1.content`,
		ContainerdRegistryConfigPath: `C:\ProgramData\containerd\certs.d`,
		PersistenceHostPath:          `C:\ProgramData\spegel`,
		BasicAuthSecretsDir:          `C:\spegel\basic-auth`,
	}
	if d != want {
		t.Errorf("Defaults() = %+v, want %+v", d, want)
	}
}
