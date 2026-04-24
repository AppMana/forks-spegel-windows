//go:build linux

package version

import (
	"os"
	"strings"
)

func getDistro() string {
	unknownDistro := "unknown"
	b, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return unknownDistro
	}
	for line := range strings.SplitSeq(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(line[len("PRETTY_NAME="):], `"`)
		}
	}
	return unknownDistro
}
