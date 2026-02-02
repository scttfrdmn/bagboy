package init

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type ProjectInfo struct {
	Name        string
	Version     string
	Description string
	Author      string
	Homepage    string
	License     string
	GitHubOwner string
	GitHubRepo  string
	Language    string
	Binaries    map[string]string
}

func DetectProject() (*ProjectInfo, error) {
	info := &ProjectInfo{
		Binaries: make(map[string]string),
	}

	// Try to detect from various project files
	if err := detectFromGo(info); err == nil {
		info.Language = "go"
	} else if err := detectFromNodeJS(info); err == nil {
		info.Language = "nodejs"
	} else if err := detectFromRust(info); err == nil {
		info.Language = "rust"
	} else if err := detectFromPython(info); err == nil {
		info.Language = "python"
	}

	// Try to get git info
	detectFromGit(info)

	// Try to find existing binaries
	detectBinaries(info)

	// Set defaults if not detected
	if info.Name == "" {
		if cwd, err := os.Getwd(); err == nil {
			info.Name = filepath.Base(cwd)
		}
	}
	if info.Version == "" {
		info.Version = "0.1.0"
	}
	if info.License == "" {
		info.License = "MIT"
	}

	return info, nil
}

func detectFromGo(info *ProjectInfo) error {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			module := strings.TrimPrefix(line, "module ")
			parts := strings.Split(module, "/")
			if len(parts) > 0 {
				info.Name = parts[len(parts)-1]
			}
			if strings.Contains(module, "github.com") && len(parts) >= 3 {
				info.GitHubOwner = parts[1]
				info.GitHubRepo = parts[2]
			}
			break
		}
	}

	return nil
}

func detectFromNodeJS(info *ProjectInfo) error {
	data, err := os.ReadFile("package.json")
	if err != nil {
		return err
	}

	var pkg struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
		Author      string `json:"author"`
		License     string `json:"license"`
		Homepage    string `json:"homepage"`
		Repository  struct {
			URL string `json:"url"`
		} `json:"repository"`
	}

	if err := yaml.Unmarshal(data, &pkg); err != nil {
		return err
	}

	info.Name = pkg.Name
	info.Version = pkg.Version
	info.Description = pkg.Description
	info.Author = pkg.Author
	info.License = pkg.License
	info.Homepage = pkg.Homepage

	if strings.Contains(pkg.Repository.URL, "github.com") {
		re := regexp.MustCompile(`github\.com[:/]([^/]+)/([^/]+)`)
		matches := re.FindStringSubmatch(pkg.Repository.URL)
		if len(matches) >= 3 {
			info.GitHubOwner = matches[1]
			info.GitHubRepo = strings.TrimSuffix(matches[2], ".git")
		}
	}

	return nil
}

func detectFromRust(info *ProjectInfo) error {
	data, err := os.ReadFile("Cargo.toml")
	if err != nil {
		return err
	}

	// Simple TOML parsing for basic fields
	lines := strings.Split(string(data), "\n")
	inPackage := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "[package]" {
			inPackage = true
			continue
		}

		if strings.HasPrefix(line, "[") && line != "[package]" {
			inPackage = false
			continue
		}

		if !inPackage {
			continue
		}

		if strings.HasPrefix(line, "name = ") {
			info.Name = strings.Trim(strings.TrimPrefix(line, "name = "), `"`)
		} else if strings.HasPrefix(line, "version = ") {
			info.Version = strings.Trim(strings.TrimPrefix(line, "version = "), `"`)
		} else if strings.HasPrefix(line, "description = ") {
			info.Description = strings.Trim(strings.TrimPrefix(line, "description = "), `"`)
		} else if strings.HasPrefix(line, "license = ") {
			info.License = strings.Trim(strings.TrimPrefix(line, "license = "), `"`)
		}
	}

	return nil
}

func detectFromPython(info *ProjectInfo) error {
	data, err := os.ReadFile("pyproject.toml")
	if err != nil {
		return err
	}

	// Simple TOML parsing for basic fields
	lines := strings.Split(string(data), "\n")
	inProject := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "[project]" {
			inProject = true
			continue
		}

		if strings.HasPrefix(line, "[") && line != "[project]" {
			inProject = false
			continue
		}

		if !inProject {
			continue
		}

		if strings.HasPrefix(line, "name = ") {
			info.Name = strings.Trim(strings.TrimPrefix(line, "name = "), `"`)
		} else if strings.HasPrefix(line, "version = ") {
			info.Version = strings.Trim(strings.TrimPrefix(line, "version = "), `"`)
		} else if strings.HasPrefix(line, "description = ") {
			info.Description = strings.Trim(strings.TrimPrefix(line, "description = "), `"`)
		}
	}

	return nil
}

func detectFromGit(info *ProjectInfo) {
	// Try to get remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	url := strings.TrimSpace(string(output))
	if strings.Contains(url, "github.com") {
		re := regexp.MustCompile(`github\.com[:/]([^/]+)/([^/]+)`)
		matches := re.FindStringSubmatch(url)
		if len(matches) >= 3 {
			info.GitHubOwner = matches[1]
			info.GitHubRepo = strings.TrimSuffix(matches[2], ".git")
			if info.Homepage == "" {
				info.Homepage = fmt.Sprintf("https://github.com/%s/%s", info.GitHubOwner, info.GitHubRepo)
			}
		}
	}
}

func detectBinaries(info *ProjectInfo) {
	// Common binary locations
	locations := []string{
		"dist",
		"build",
		"target/release",
		"bin",
	}

	platforms := []string{
		"darwin-amd64",
		"darwin-arm64",
		"linux-amd64",
		"linux-arm64",
		"windows-amd64",
	}

	for _, location := range locations {
		if _, err := os.Stat(location); os.IsNotExist(err) {
			continue
		}

		for _, platform := range platforms {
			// Try various naming patterns
			patterns := []string{
				filepath.Join(location, info.Name+"-"+platform),
				filepath.Join(location, info.Name+"-"+platform+".exe"),
				filepath.Join(location, platform, info.Name),
				filepath.Join(location, platform, info.Name+".exe"),
			}

			for _, pattern := range patterns {
				if _, err := os.Stat(pattern); err == nil {
					info.Binaries[platform] = pattern
					break
				}
			}
		}
	}
}

func PromptUser(info *ProjectInfo) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Project name [%s]: ", info.Name)
	if input := readLine(reader); input != "" {
		info.Name = input
	}

	fmt.Printf("Version [%s]: ", info.Version)
	if input := readLine(reader); input != "" {
		info.Version = input
	}

	fmt.Printf("Description [%s]: ", info.Description)
	if input := readLine(reader); input != "" {
		info.Description = input
	}

	fmt.Printf("Author [%s]: ", info.Author)
	if input := readLine(reader); input != "" {
		info.Author = input
	}

	fmt.Printf("Homepage [%s]: ", info.Homepage)
	if input := readLine(reader); input != "" {
		info.Homepage = input
	}

	fmt.Printf("License [%s]: ", info.License)
	if input := readLine(reader); input != "" {
		info.License = input
	}

	if info.GitHubOwner != "" {
		fmt.Printf("GitHub owner [%s]: ", info.GitHubOwner)
		if input := readLine(reader); input != "" {
			info.GitHubOwner = input
		}
	}

	if info.GitHubRepo != "" {
		fmt.Printf("GitHub repo [%s]: ", info.GitHubRepo)
		if input := readLine(reader); input != "" {
			info.GitHubRepo = input
		}
	}

	return nil
}

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
