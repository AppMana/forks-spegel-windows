//go:build windows

package version

import "golang.org/x/sys/windows/registry"

func getDistro() string {
	unknownDistro := "unknown"
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return unknownDistro
	}
	defer k.Close()
	name, _, err := k.GetStringValue("ProductName")
	if err != nil {
		return unknownDistro
	}
	return name
}
