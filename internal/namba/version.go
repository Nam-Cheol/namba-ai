package namba

import "strings"

const devVersionLabel = "dev"

// cliVersion is injected at build time for tagged release binaries.
var cliVersion = devVersionLabel

func Version() string {
	return resolveVersion(cliVersion)
}

func VersionLine() string {
	return formatVersionLine(Version())
}

func resolveVersion(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return devVersionLabel
	}
	return trimmed
}

func formatVersionLine(version string) string {
	return "namba " + resolveVersion(version)
}
