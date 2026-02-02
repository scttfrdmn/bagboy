package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

type Packager struct{}

func New() *Packager {
	return &Packager{}
}

func (p *Packager) Name() string {
	return "docker"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Description == "" {
		return fmt.Errorf("description is required for Docker image")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	dockerDir := filepath.Join("dist", "docker")
	if err := os.MkdirAll(dockerDir, 0755); err != nil {
		return "", err
	}

	// Create Dockerfile
	dockerfilePath := filepath.Join(dockerDir, "Dockerfile")
	if err := p.createDockerfile(dockerfilePath, cfg); err != nil {
		return "", err
	}

	// Create .dockerignore
	dockerignorePath := filepath.Join(dockerDir, ".dockerignore")
	dockerignoreContent := `# Ignore everything except binaries
*
!dist/
dist/*
!dist/*-linux-*
*.log
*.tmp
.git/
.github/
node_modules/
target/debug/
*.deb
*.rpm
*.AppImage`
	if err := os.WriteFile(dockerignorePath, []byte(dockerignoreContent), 0644); err != nil {
		return "", err
	}

	// Create docker-compose.yml for easy deployment
	composePath := filepath.Join(dockerDir, "docker-compose.yml")
	if err := p.createDockerCompose(composePath, cfg); err != nil {
		return "", err
	}

	// Create build script
	buildScriptPath := filepath.Join(dockerDir, "build.sh")
	if err := p.createBuildScript(buildScriptPath, cfg); err != nil {
		return "", err
	}

	return dockerDir, nil
}

func (p *Packager) createDockerfile(path string, cfg *config.Config) error {
	// Find Linux binary
	var linuxBinary string
	for arch, binaryPath := range cfg.Binaries {
		if strings.HasPrefix(arch, "linux-") {
			linuxBinary = binaryPath
			break
		}
	}
	if linuxBinary == "" {
		return fmt.Errorf("no Linux binary found for Docker image")
	}

	tmpl := `# Multi-stage build for {{.Name}}
FROM alpine:latest as builder

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary
COPY {{.BinaryPath}} /root/{{.Name}}
RUN chmod +x /root/{{.Name}}

# Final stage - minimal image
FROM scratch

# Copy ca-certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /root/{{.Name}} /{{.Name}}

# Metadata
LABEL maintainer="{{.Author}}"
LABEL description="{{.Description}}"
LABEL version="{{.Version}}"
LABEL homepage="{{.Homepage}}"
LABEL org.opencontainers.image.source="{{.Homepage}}"
LABEL org.opencontainers.image.description="{{.Description}}"
LABEL org.opencontainers.image.version="{{.Version}}"

# Set the binary as entrypoint
ENTRYPOINT ["/{{.Name}}"]
CMD ["--help"]`

	t, err := template.New("dockerfile").Parse(tmpl)
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
		BinaryPath string
	}{
		Config:     cfg,
		BinaryPath: linuxBinary,
	}

	return t.Execute(f, data)
}

func (p *Packager) createDockerCompose(path string, cfg *config.Config) error {
	tmpl := `version: '3.8'

services:
  {{.Name}}:
    build: .
    image: {{.ImageName}}:{{.Version}}
    container_name: {{.Name}}
    restart: unless-stopped
    # Uncomment and modify as needed:
    # ports:
    #   - "8080:8080"
    # volumes:
    #   - ./data:/data
    # environment:
    #   - ENV_VAR=value

# For CLI usage:
# docker-compose run --rm {{.Name}} [command]`

	t, err := template.New("compose").Parse(tmpl)
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
		ImageName string
	}{
		Config:    cfg,
		ImageName: strings.ToLower(cfg.Name),
	}

	return t.Execute(f, data)
}

func (p *Packager) createBuildScript(path string, cfg *config.Config) error {
	tmpl := `#!/bin/bash
set -e

# Build script for {{.Name}} Docker image

IMAGE_NAME="{{.ImageName}}"
VERSION="{{.Version}}"
LATEST_TAG="${IMAGE_NAME}:latest"
VERSION_TAG="${IMAGE_NAME}:${VERSION}"

echo "Building Docker image for {{.Name}} v${VERSION}..."

# Build the image
docker build -t "${VERSION_TAG}" -t "${LATEST_TAG}" .

echo "✅ Built Docker images:"
echo "  ${VERSION_TAG}"
echo "  ${LATEST_TAG}"

echo ""
echo "Usage:"
echo "  docker run --rm ${LATEST_TAG} --help"
echo "  docker run --rm ${LATEST_TAG} [command]"
echo ""
echo "To push to registry:"
echo "  docker push ${VERSION_TAG}"
echo "  docker push ${LATEST_TAG}"`

	t, err := template.New("build").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := os.Chmod(path, 0755); err != nil {
		return err
	}

	data := struct {
		*config.Config
		ImageName string
	}{
		Config:    cfg,
		ImageName: strings.ToLower(cfg.Name),
	}

	return t.Execute(f, data)
}
