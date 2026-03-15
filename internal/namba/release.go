package namba

import "fmt"

type releaseTarget struct {
	GOOS      string
	GOARCH    string
	Archive   string
	AssetName string
}

func releaseTargets() []releaseTarget {
	return []releaseTarget{
		{GOOS: "windows", GOARCH: "amd64", Archive: "zip", AssetName: "namba_Windows_x86_64.zip"},
		{GOOS: "windows", GOARCH: "arm64", Archive: "zip", AssetName: "namba_Windows_arm64.zip"},
		{GOOS: "linux", GOARCH: "amd64", Archive: "tar.gz", AssetName: "namba_Linux_x86_64.tar.gz"},
		{GOOS: "linux", GOARCH: "arm64", Archive: "tar.gz", AssetName: "namba_Linux_arm64.tar.gz"},
		{GOOS: "darwin", GOARCH: "amd64", Archive: "tar.gz", AssetName: "namba_macOS_x86_64.tar.gz"},
		{GOOS: "darwin", GOARCH: "arm64", Archive: "tar.gz", AssetName: "namba_macOS_arm64.tar.gz"},
	}
}

func releaseAssetName(goos, goarch string) (string, error) {
	for _, target := range releaseTargets() {
		if target.GOOS == goos && target.GOARCH == goarch {
			return target.AssetName, nil
		}
	}
	return "", fmt.Errorf("unsupported release target %s/%s", goos, goarch)
}
