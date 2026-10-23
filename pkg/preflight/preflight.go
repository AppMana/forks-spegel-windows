package preflight

import (
	"fmt"
	"os"

	"github.com/kvick-org/pkg/version"
)

const (
	KubernetesEnvVar = "KUBERNETES_SERVICE_HOST"
)

func Check(versionInfo version.Info) error {
	if os.Getenv(KubernetesEnvVar) == "" {
		return nil
	}
	if versionInfo.Runtime.OS == "linux" && versionInfo.Runtime.Distro == "Distroless" {
		return nil
	}
	// HostProcess containers execute on the Windows host, outside the image's
	// userspace. Kubelet provides this path to the extracted container files.
	if versionInfo.Runtime.OS == "windows" && os.Getenv("CONTAINER_SANDBOX_MOUNT_POINT") != "" {
		return nil
	}
	return fmt.Errorf("unsupported container distro %s", versionInfo.Runtime.Distro)
}
