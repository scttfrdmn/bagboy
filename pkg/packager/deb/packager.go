package deb

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/blakesmith/ar"
	"github.com/scttfrdmn/bagboy/pkg/config"
)

type Packager struct{}

func New() *Packager {
	return &Packager{}
}

func (p *Packager) Name() string {
	return "deb"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Packages.Deb.Maintainer == "" {
		return fmt.Errorf("deb.maintainer is required")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	// Create temp directory for package structure
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("%s-deb-%s", cfg.Name, cfg.Version))
	if err := os.RemoveAll(tempDir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	// Create DEBIAN directory
	debianDir := filepath.Join(tempDir, "DEBIAN")
	if err := os.MkdirAll(debianDir, 0755); err != nil {
		return "", err
	}

	// Create control file
	controlPath := filepath.Join(debianDir, "control")
	if err := p.createControlFile(controlPath, cfg); err != nil {
		return "", err
	}

	// Create binary directory and copy binary
	binDir := filepath.Join(tempDir, "usr", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", err
	}

	// Find Linux binary
	var linuxBinary string
	for arch, path := range cfg.Binaries {
		if strings.HasPrefix(arch, "linux-") {
			linuxBinary = path
			break
		}
	}

	if linuxBinary == "" {
		return "", fmt.Errorf("no Linux binary found")
	}

	// Copy binary
	src, err := os.Open(linuxBinary)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(filepath.Join(binDir, cfg.Name))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	if err := os.Chmod(filepath.Join(binDir, cfg.Name), 0755); err != nil {
		return "", err
	}

	// Create the .deb package
	outputPath := filepath.Join("dist", fmt.Sprintf("%s_%s_amd64.deb", cfg.Name, cfg.Version))
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", err
	}

	return outputPath, p.createDebPackage(tempDir, outputPath)
}

func (p *Packager) createControlFile(path string, cfg *config.Config) error {
	tmpl := `Package: {{.Name}}
Version: {{.Version}}
Section: {{.Section}}
Priority: {{.Priority}}
Architecture: amd64
Maintainer: {{.Maintainer}}
Description: {{.Description}}
Homepage: {{.Homepage}}`

	t, err := template.New("control").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	data := struct {
		*config.Config
		Section    string
		Priority   string
		Maintainer string
	}{
		Config:     cfg,
		Section:    cfg.Packages.Deb.Section,
		Priority:   cfg.Packages.Deb.Priority,
		Maintainer: cfg.Packages.Deb.Maintainer,
	}

	if data.Section == "" {
		data.Section = "utils"
	}
	if data.Priority == "" {
		data.Priority = "optional"
	}

	return t.Execute(f, data)
}

func (p *Packager) createDebPackage(sourceDir, outputPath string) error {
	// Create data.tar.gz
	dataPath := filepath.Join(os.TempDir(), "data.tar.gz")
	if err := p.createTarGz(sourceDir, dataPath, []string{"DEBIAN"}); err != nil {
		return err
	}
	defer os.Remove(dataPath)

	// Create control.tar.gz
	controlPath := filepath.Join(os.TempDir(), "control.tar.gz")
	debianDir := filepath.Join(sourceDir, "DEBIAN")
	if err := p.createTarGz(debianDir, controlPath, nil); err != nil {
		return err
	}
	defer os.Remove(controlPath)

	// Create debian-binary
	debianBinaryPath := filepath.Join(os.TempDir(), "debian-binary")
	if err := os.WriteFile(debianBinaryPath, []byte("2.0\n"), 0644); err != nil {
		return err
	}
	defer os.Remove(debianBinaryPath)

	// Create .deb file using ar
	debFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer debFile.Close()

	arWriter := ar.NewWriter(debFile)

	// Add files to ar archive
	files := []string{debianBinaryPath, controlPath, dataPath}
	names := []string{"debian-binary", "control.tar.gz", "data.tar.gz"}

	for i, file := range files {
		if err := p.addFileToAr(arWriter, file, names[i]); err != nil {
			return err
		}
	}

	return nil
}

func (p *Packager) createTarGz(sourceDir, outputPath string, exclude []string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		relPath, _ := filepath.Rel(sourceDir, path)
		for _, ex := range exclude {
			if strings.HasPrefix(relPath, ex) {
				return nil
			}
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		header.Name = relPath
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			return err
		}

		return nil
	})
}

func (p *Packager) addFileToAr(arWriter *ar.Writer, filePath, name string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header := &ar.Header{
		Name:    name,
		Size:    info.Size(),
		Mode:    int64(info.Mode()),
		ModTime: info.ModTime(),
	}

	if err := arWriter.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(arWriter, file)
	return err
}
