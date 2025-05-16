// setup_osquery.go
package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func SetupOsquery() error {
	binDir := filepath.Join(".", "bin")
	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "linux":
		return downloadLinuxOsquery(binDir)
	case "windows":
		return downloadWindowsOsquery(binDir)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func downloadLinuxOsquery(binDir string) error {
	url := "https://pkg.osquery.io/linux/osquery-5.9.1_1.linux_x86_64.tar.gz"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if filepath.Base(hdr.Name) == "osqueryi" {
			outPath := filepath.Join(binDir, "osqueryi")
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, tr)
			if err != nil {
				return err
			}
			err = os.Chmod(outPath, 0755)
			return err
		}
	}
	return fmt.Errorf("osqueryi binary not found in archive")
}

func downloadWindowsOsquery(binDir string) error {
	url := "https://osquery-packages.s3.amazonaws.com/windows/osquery-5.9.1.msi"
	outPath := filepath.Join(binDir, "osquery.msi")
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return err
	}

	// Use an absolute path for extraction
	extractDir, err := filepath.Abs(filepath.Join(binDir, "osquery_extracted"))
	if err != nil {
		return err
	}

	// Extract osqueryi.exe from .msi (requires msiexec)
	cmd := exec.Command("msiexec", "/a", outPath, "/qn", "TARGETDIR="+extractDir)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to extract MSI: %v", err)
	}

	// Search for osqueryi.exe in the extracted directory
	var src string
	_ = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Name() == "osqueryi.exe" {
			src = path
			return io.EOF // stop walking
		}
		return nil
	})
	if src == "" {
		return fmt.Errorf("osqueryi.exe not found after MSI extraction")
	}

	dst := filepath.Join(binDir, "osqueryi.exe")
	err = os.Rename(src, dst)
	if err != nil {
		return fmt.Errorf("failed to move osqueryi.exe: %v", err)
	}

	return nil
}
