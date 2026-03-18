package namba

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const updateRepo = "Nam-Cheol/namba-ai"

type updateOptions struct {
	Version string
}

func (a *App) runUpdate(ctx context.Context, args []string) error {
	opts, err := parseUpdateArgs(args)
	if err != nil {
		return err
	}

	assetName, err := releaseAssetName(a.goos, a.goarch)
	if err != nil {
		return err
	}

	execPath, err := a.executablePath()
	if err != nil {
		return fmt.Errorf("resolve current executable: %w", err)
	}

	url := releaseDownloadURL(opts.Version, assetName)
	archiveData, err := a.downloadURL(ctx, url)
	if err != nil {
		return formatUpdateDownloadError(err, url, opts.Version, assetName)
	}

	checksumsURL := releaseDownloadURL(opts.Version, "checksums.txt")
	checksumsData, err := a.downloadURL(ctx, checksumsURL)
	if err != nil {
		return formatUpdateDownloadError(err, checksumsURL, opts.Version, "checksums.txt")
	}
	if err := verifyUpdateChecksum(assetName, archiveData, checksumsData); err != nil {
		return err
	}

	binary, err := extractUpdatedBinary(assetName, archiveData)
	if err != nil {
		return err
	}

	if a.goos == "windows" {
		if err := a.scheduleWindowsUpdate(execPath, binary); err != nil {
			return err
		}
		fmt.Fprintf(a.stdout, "Scheduled NambaAI update to %s. Restart the terminal after this command exits.\n", updateVersionLabel(opts.Version))
		return nil
	}

	if err := writeUpdatedExecutable(execPath, binary); err != nil {
		return err
	}

	fmt.Fprintf(a.stdout, "Updated NambaAI to %s at %s\n", updateVersionLabel(opts.Version), execPath)
	return nil
}

func parseUpdateArgs(args []string) (updateOptions, error) {
	opts := updateOptions{Version: "latest"}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--version":
			i++
			if i >= len(args) {
				return updateOptions{}, errors.New("--version requires a value")
			}
			opts.Version = args[i]
		default:
			return updateOptions{}, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return opts, nil
}

func releaseDownloadURL(version, assetName string) string {
	if strings.TrimSpace(version) == "" || version == "latest" {
		return fmt.Sprintf("https://github.com/%s/releases/latest/download/%s", updateRepo, assetName)
	}
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", updateRepo, version, assetName)
}

func updateVersionLabel(version string) string {
	if strings.TrimSpace(version) == "" {
		return "latest"
	}
	return version
}

func formatUpdateDownloadError(err error, url, version, assetName string) error {
	message := err.Error()
	if strings.Contains(message, "status 404") {
		if version == "latest" {
			return fmt.Errorf("failed to download %s (404). No GitHub Release has been published yet, or the latest release does not contain %s", url, assetName)
		}
		return fmt.Errorf("failed to download %s (404). Release %q was not found, or it does not contain %s", url, version, assetName)
	}
	return fmt.Errorf("download update archive: %w", err)
}

func verifyUpdateChecksum(assetName string, archiveData, checksumsData []byte) error {
	expected, err := parseChecksumManifest(assetName, checksumsData)
	if err != nil {
		return err
	}

	actualSum := sha256.Sum256(archiveData)
	actual := hex.EncodeToString(actualSum[:])
	if actual != expected {
		return fmt.Errorf("checksum mismatch for %s", assetName)
	}

	return nil
}

func parseChecksumManifest(assetName string, checksumsData []byte) (string, error) {
	for _, line := range strings.Split(string(checksumsData), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 {
			continue
		}

		name := strings.TrimPrefix(fields[len(fields)-1], "*")
		if name != assetName {
			continue
		}

		sum := strings.ToLower(fields[0])
		if len(sum) != 64 {
			return "", fmt.Errorf("invalid checksum for %s", assetName)
		}
		if _, err := hex.DecodeString(sum); err != nil {
			return "", fmt.Errorf("invalid checksum for %s", assetName)
		}
		return sum, nil
	}

	return "", fmt.Errorf("checksums.txt does not contain %s", assetName)
}

func extractUpdatedBinary(assetName string, archiveData []byte) ([]byte, error) {
	switch {
	case strings.HasSuffix(assetName, ".zip"):
		return extractUpdatedBinaryFromZip(archiveData)
	case strings.HasSuffix(assetName, ".tar.gz"):
		return extractUpdatedBinaryFromTarGz(archiveData)
	default:
		return nil, fmt.Errorf("unsupported release archive %q", assetName)
	}
}

func extractUpdatedBinaryFromZip(archiveData []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
	if err != nil {
		return nil, fmt.Errorf("open zip archive: %w", err)
	}

	for _, file := range reader.File {
		if filepath.Base(file.Name) != "namba.exe" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open %s in zip archive: %w", file.Name, err)
		}
		defer rc.Close()
		return readAllBytes(rc)
	}

	return nil, errors.New("namba.exe was not found in the downloaded archive")
}

func extractUpdatedBinaryFromTarGz(archiveData []byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(archiveData))
	if err != nil {
		return nil, fmt.Errorf("open tar.gz archive: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read tar.gz archive: %w", err)
		}
		if filepath.Base(header.Name) != "namba" {
			continue
		}
		return readAllBytes(tarReader)
	}

	return nil, errors.New("namba was not found in the downloaded archive")
}

func readAllBytes(reader io.Reader) ([]byte, error) {
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(reader); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func writeUpdatedExecutable(path string, binary []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat current executable: %w", err)
	}

	tempPath := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".tmp")
	if err := os.WriteFile(tempPath, binary, info.Mode().Perm()); err != nil {
		return fmt.Errorf("write updated executable: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		if removeErr := os.Remove(path); removeErr != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("replace executable: %w", err)
		}
		if retryErr := os.Rename(tempPath, path); retryErr != nil {
			if runtime.GOOS == "windows" {
				writeErr := os.WriteFile(path, binary, info.Mode().Perm())
				_ = os.Remove(tempPath)
				if writeErr == nil {
					return nil
				}
			}
			_ = os.Remove(tempPath)
			return fmt.Errorf("replace executable after removing old binary: %w", retryErr)
		}
	}
	return nil
}

func (a *App) scheduleWindowsUpdate(execPath string, binary []byte) error {
	tempRoot, err := os.MkdirTemp("", "namba-update-*")
	if err != nil {
		return fmt.Errorf("create update temp dir: %w", err)
	}

	sourcePath := filepath.Join(tempRoot, "namba.exe")
	if err := os.WriteFile(sourcePath, binary, 0o755); err != nil {
		_ = os.RemoveAll(tempRoot)
		return fmt.Errorf("write downloaded executable: %w", err)
	}

	scriptPath := filepath.Join(tempRoot, "apply-update.ps1")
	if err := os.WriteFile(scriptPath, []byte(buildWindowsUpdateScript(os.Getpid(), sourcePath, execPath, tempRoot)), 0o644); err != nil {
		_ = os.RemoveAll(tempRoot)
		return fmt.Errorf("write update helper script: %w", err)
	}

	if err := a.startCmd("powershell", []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath}, ""); err != nil {
		_ = os.RemoveAll(tempRoot)
		return fmt.Errorf("start windows update helper: %w", err)
	}
	return nil
}

func buildWindowsUpdateScript(pid int, sourcePath, targetPath, cleanupRoot string) string {
	return strings.Join([]string{
		"$ErrorActionPreference = 'Stop'",
		fmt.Sprintf("$targetPid = %d", pid),
		fmt.Sprintf("$sourcePath = '%s'", quotePowerShellLiteral(sourcePath)),
		fmt.Sprintf("$targetPath = '%s'", quotePowerShellLiteral(targetPath)),
		fmt.Sprintf("$cleanupRoot = '%s'", quotePowerShellLiteral(cleanupRoot)),
		"while (Get-Process -Id $targetPid -ErrorAction SilentlyContinue) {",
		"  Start-Sleep -Milliseconds 200",
		"}",
		"Copy-Item -Path $sourcePath -Destination $targetPath -Force",
		"Remove-Item -Path $cleanupRoot -Recurse -Force -ErrorAction SilentlyContinue",
	}, "\n") + "\n"
}

func quotePowerShellLiteral(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
