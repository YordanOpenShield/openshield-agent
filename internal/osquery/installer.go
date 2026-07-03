package osquery

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	osqueryVersion = "5.12.1"
)

var bundledOsqueryPath string

func init() {
	// Set the bundled osquery path based on the agent's location
	exePath, err := os.Executable()
	if err != nil {
		exePath = os.Args[0]
	}
	baseDir := filepath.Dir(exePath)
	if runtime.GOOS == "windows" {
		bundledOsqueryPath = filepath.Join(baseDir, "bin", "osqueryi.exe")
	} else {
		bundledOsqueryPath = filepath.Join(baseDir, "bin", "osqueryi")
	}
}

// OsqueryPath returns the path to the osquery binary
func OsqueryPath() string {
	return bundledOsqueryPath
}

// EnsureOsquery ensures osquery is available, installing it if necessary
func EnsureOsquery() error {
	// 1. Check if osquery is in system PATH
	if path, err := exec.LookPath("osqueryi"); err == nil {
		bundledOsqueryPath = path
		return nil
	}

	// 2. Check if osquery is bundled with the agent
	if _, err := os.Stat(bundledOsqueryPath); err == nil {
		return testOsquery(bundledOsqueryPath)
	}

	// 3. Auto-download osquery
	return downloadOsquery()
}

// testOsquery verifies the binary works
func testOsquery(path string) error {
	cmd := exec.Command(path, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("osquery binary test failed: %v, output: %s", err, string(output))
	}
	return nil
}

// downloadOsquery downloads and extracts osquery for the current platform
func downloadOsquery() error {
	binDir := filepath.Dir(bundledOsqueryPath)
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	url, filename := getDownloadURL()
	archivePath := filepath.Join(binDir, filename)

	fmt.Printf("Downloading osquery from %s...\n", url)
	if err := downloadFile(url, archivePath); err != nil {
		return fmt.Errorf("failed to download osquery: %w", err)
	}
	defer os.Remove(archivePath)

	fmt.Println("Extracting osquery...")
	if err := extractArchive(archivePath, binDir); err != nil {
		return fmt.Errorf("failed to extract osquery: %w", err)
	}

	// Make binary executable on Unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(bundledOsqueryPath, 0755); err != nil {
			return fmt.Errorf("failed to make osquery executable: %w", err)
		}
	}

	return testOsquery(bundledOsqueryPath)
}

// getDownloadURL returns the download URL and filename for the current platform
func getDownloadURL() (string, string) {
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}

	switch runtime.GOOS {
	case "linux":
		filename := fmt.Sprintf("osquery-%s_1.linux_%s.tar.gz", osqueryVersion, arch)
		return fmt.Sprintf("https://pkg.osquery.io/linux/%s", filename), filename
	case "darwin":
		filename := fmt.Sprintf("osquery-%s.pkg", osqueryVersion)
		return fmt.Sprintf("https://pkg.osquery.io/darwin/%s", filename), filename
	case "windows":
		// Use GitHub releases for Windows - more reliable
		filename := fmt.Sprintf("osquery-%s.windows_x86_64.zip", osqueryVersion)
		return fmt.Sprintf("https://github.com/osquery/osquery/releases/download/%s/%s", osqueryVersion, filename), filename
	default:
		return "", ""
	}
}

// downloadFile downloads a file from URL to the specified path
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractArchive extracts the downloaded archive based on file extension
func extractArchive(archivePath, destDir string) error {
	ext := filepath.Ext(archivePath)
	if ext == ".zip" {
		return extractZip(archivePath, destDir)
	}
	if ext == ".gz" || ext == ".tgz" {
		return extractTarGz(archivePath, destDir)
	}
	return fmt.Errorf("unsupported archive format: %s", ext)
}

// extractTarGz extracts a .tar.gz archive
func extractTarGz(gzipPath, destDir string) error {
	file, err := os.Open(gzipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Only extract osqueryi binary
		if header.Typeflag == tar.TypeReg {
			if filepath.Base(header.Name) == "osqueryi" || filepath.Base(header.Name) == "osqueryi.exe" {
				outPath := filepath.Join(destDir, filepath.Base(header.Name))
				outFile, err := os.Create(outPath)
				if err != nil {
					return err
				}
				if _, err := io.Copy(outFile, tarReader); err != nil {
					outFile.Close()
					return err
				}
				outFile.Close()
			}
		}
	}
	return nil
}

// extractZip extracts a .zip archive
func extractZip(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		// Only extract osqueryi.exe
		if filepath.Base(file.Name) == "osqueryi.exe" {
			outPath := filepath.Join(destDir, "osqueryi.exe")
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}

			rc, err := file.Open()
			if err != nil {
				outFile.Close()
				return err
			}

			_, err = io.Copy(outFile, rc)
			rc.Close()
			outFile.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
