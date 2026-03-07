package version

import (
	"runtime/debug"
	"strings"
)

var buildVersion string

func Current() string {
	if version := normalize(buildVersion); version != "" {
		return version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if version := normalize(info.Main.Version); version != "" {
			return version
		}
	}

	return "dev"
}

func normalize(raw string) string {
	version := strings.TrimSpace(raw)
	if version == "" || version == "(devel)" {
		return ""
	}
	return strings.TrimPrefix(version, "v")
}
