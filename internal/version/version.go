package version

import (
	"errors"
	"runtime/debug"
	"strconv"
)

type Info struct {
	Build   Build   `json:"build" text:"Build"`
	Runtime Runtime `json:"runtime" text:"Runtime"`
}

type Build struct {
	GoVersion string `json:"goVersion"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
}

type Runtime struct {
	Distro string `json:"distro"`
	OS     string `json:"os"`
	Arch   string `json:"arch"`
}

// Preflight is a best-effort sanity check run on startup. It used to gate
// Spegel to Linux+Distroless only; that gate has been removed to support
// Windows HostProcess deployments where the runtime is Windows Server and
// there is no "Distroless" concept. The function is kept so existing callers
// keep compiling, and so future platform-specific guards can be added without
// changing the call site in main.go. Today it performs no validation.
func (i Info) Preflight() error {
	return nil
}

func Load() (Info, error) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return Info{}, errors.New("could not read build info")
	}
	info := Info{
		Build: Build{
			GoVersion: bi.GoVersion,
		},
		Runtime: Runtime{
			Distro: getDistro(),
		},
	}
	for _, s := range bi.Settings {
		switch s.Key {
		case "GOARCH":
			info.Runtime.Arch = s.Value
		case "GOOS":
			info.Runtime.OS = s.Value
		case "vcs.revision":
			info.Build.Commit = s.Value
		case "vcs.modified":
			modified, err := strconv.ParseBool(s.Value)
			if err != nil {
				return Info{}, err
			}
			info.Build.Version = getVersion(bi.Main.Version, modified)
		}
	}
	return info, nil
}

func getVersion(mainVersion string, modified bool) string {
	develVersion := "devel"
	if modified {
		return develVersion
	}
	if mainVersion == "" {
		return develVersion
	}
	return mainVersion
}
