package internal

import "runtime/debug"

func GetModuleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "0.0.0"
	}

	for _, d := range info.Deps {
		if d.Path == "github.com/graphmetrics/graphmetrics-go" {
			return d.Version[1:]
		}
	}

	return "0.0.0"
}
