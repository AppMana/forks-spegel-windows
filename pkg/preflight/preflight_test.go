package preflight

import (
	"testing"

	"github.com/go-openapi/testify/v2/require"

	"github.com/kvick-org/pkg/version"
)

func TestCheck(t *testing.T) {
	versionInfo := version.Info{
		Runtime: version.Runtime{
			OS:     "linux",
			Distro: "ubuntu",
		},
	}
	err := Check(versionInfo)
	require.NoError(t, err)

	t.Setenv(KubernetesEnvVar, "foo")
	err = Check(versionInfo)
	require.EqualError(t, err, "unsupported container distro ubuntu")

	versionInfo.Runtime.Distro = "Distroless"
	err = Check(versionInfo)
	require.NoError(t, err)
}

func TestWindowsHostProcess(t *testing.T) {
	t.Setenv(KubernetesEnvVar, "kubernetes")
	t.Setenv("CONTAINER_SANDBOX_MOUNT_POINT", "")
	info := version.Info{Runtime: version.Runtime{OS: "windows", Distro: "unknown"}}
	require.Error(t, Check(info))
	t.Setenv("CONTAINER_SANDBOX_MOUNT_POINT", `C:\var\lib\kubelet\pods\fixture`)
	require.NoError(t, Check(info))
	info.Runtime.OS = "linux"
	require.Error(t, Check(info))
}
