package rpm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/scttfrdmn/bagboy/pkg/config"
	"github.com/scttfrdmn/bagboy/pkg/vm"
)

type Packager struct{}

func New() *Packager {
	return &Packager{}
}

func (p *Packager) Name() string {
	return "rpm"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Packages.RPM.Vendor == "" {
		return fmt.Errorf("rpm.vendor is required")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
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

	// Create RPM build directory structure
	buildDir := filepath.Join("dist", "rpm-build")
	dirs := []string{"BUILD", "RPMS", "SOURCES", "SPECS", "SRPMS"}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(buildDir, dir), 0755); err != nil {
			return "", fmt.Errorf("failed to create RPM directory %s: %w", dir, err)
		}
	}

	// Copy binary to SOURCES
	sourcePath := filepath.Join(buildDir, "SOURCES", cfg.Name)
	if err := p.copyFile(linuxBinary, sourcePath); err != nil {
		return "", fmt.Errorf("failed to copy binary: %w", err)
	}

	// Generate spec file
	specPath := filepath.Join(buildDir, "SPECS", cfg.Name+".spec")
	specContent := p.generateSpec(cfg, linuxBinary)
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write spec file: %w", err)
	}

	// Build RPM
	return p.buildRPM(ctx, buildDir, specPath, cfg)
}

func (p *Packager) generateSpec(cfg *config.Config, binaryPath string) string {
	tmpl := `Name:           {{.Name}}
Version:        {{.Version}}
Release:        1%{?dist}
Summary:        {{.Description}}
License:        {{.License}}
URL:            {{.Homepage}}
Source0:        %{name}-%{version}.tar.gz
BuildArch:      x86_64
Group:          {{.Group}}
Vendor:         {{.Vendor}}

%description
{{.Description}}

%prep
%setup -q

%build
# No build needed for pre-compiled binary

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/usr/bin
cp {{.BinaryName}} $RPM_BUILD_ROOT/usr/bin/{{.Name}}

%files
/usr/bin/{{.Name}}

%changelog
* $(date "+%a %b %d %Y") {{.Vendor}} - {{.Version}}-1
- Initial package`

	t, _ := template.New("spec").Parse(tmpl)

	data := struct {
		*config.Config
		Group      string
		Vendor     string
		BinaryName string
	}{
		Config:     cfg,
		Group:      cfg.Packages.RPM.Group,
		Vendor:     cfg.Packages.RPM.Vendor,
		BinaryName: filepath.Base(binaryPath),
	}

	if data.Group == "" {
		data.Group = "Applications/System"
	}

	var result strings.Builder
	t.Execute(&result, data)
	return result.String()
}

func (p *Packager) buildRPM(ctx context.Context, buildDir, specPath string, cfg *config.Config) (string, error) {
	// Check if rpmbuild is available
	if _, err := exec.LookPath("rpmbuild"); err != nil {
		// Try VM if enabled
		if cfg.VM.Enabled {
			return p.buildRPMWithVM(ctx, buildDir, specPath, cfg)
		}
		return "", fmt.Errorf("rpmbuild not found - install rpm-build package or enable VM support in bagboy.yaml")
	}

	// Build RPM natively
	cmd := exec.CommandContext(ctx, "rpmbuild",
		"--define", "_topdir "+buildDir,
		"-bb", specPath)

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("rpmbuild failed: %w\nOutput: %s", err, output)
	}

	// Find generated RPM
	rpmPattern := filepath.Join(buildDir, "RPMS", "x86_64", fmt.Sprintf("%s-%s-*.rpm", cfg.Name, cfg.Version))
	matches, err := filepath.Glob(rpmPattern)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("RPM file not found after build")
	}

	// Move to dist directory
	finalPath := filepath.Join("dist", filepath.Base(matches[0]))
	if err := os.Rename(matches[0], finalPath); err != nil {
		return "", fmt.Errorf("failed to move RPM: %w", err)
	}

	return finalPath, nil
}

func (p *Packager) buildRPMWithVM(ctx context.Context, buildDir, specPath string, cfg *config.Config) (string, error) {
	vmMgr := vm.NewManager(&vm.Config{
		Enabled:  true,
		Provider: cfg.VM.Provider,
		Docker: vm.DockerConfig{
			Image:   getVMImage(cfg),
			WorkDir: ".",
		},
	})

	if !vmMgr.IsAvailable() {
		return "", fmt.Errorf("VM provider not available - install Docker")
	}

	// Build command to run in container
	cmd := fmt.Sprintf(`
		yum install -y rpm-build rpmdevtools && \
		cd /work && \
		rpmbuild --define "_topdir %s" -bb %s
	`, buildDir, specPath)

	_, err := vmMgr.BuildInVM(ctx, getVMImage(cfg), ".", cmd)
	if err != nil {
		return "", fmt.Errorf("VM build failed: %w", err)
	}

	// Find generated RPM
	rpmPattern := filepath.Join(buildDir, "RPMS", "x86_64", fmt.Sprintf("%s-%s-*.rpm", cfg.Name, cfg.Version))
	matches, err := filepath.Glob(rpmPattern)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("RPM file not found after build")
	}

	// Move to dist directory
	finalPath := filepath.Join("dist", filepath.Base(matches[0]))
	if err := os.Rename(matches[0], finalPath); err != nil {
		return "", fmt.Errorf("failed to move RPM: %w", err)
	}

	return finalPath, nil
}

func getVMImage(cfg *config.Config) string {
	if cfg.VM.Docker.Image != "" {
		return cfg.VM.Docker.Image
	}
	return "fedora:38"
}

func (p *Packager) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0755)
}
