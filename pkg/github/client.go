package github

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/scttfrdmn/bagboy/pkg/config"
	"golang.org/x/oauth2"
)

type Client struct {
	gh  *github.Client
	cfg *config.GitHubConfig
}

func NewClient(cfg *config.GitHubConfig) (*Client, error) {
	token := os.Getenv(cfg.TokenEnv)
	if token == "" {
		return nil, fmt.Errorf("GitHub token not found in environment variable %s", cfg.TokenEnv)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	return &Client{
		gh:  github.NewClient(tc),
		cfg: cfg,
	}, nil
}

func (c *Client) CreateRelease(ctx context.Context, cfg *config.Config, assets []string) (*github.RepositoryRelease, error) {
	release := &github.RepositoryRelease{
		TagName:              github.String("v" + cfg.Version),
		Name:                 github.String("v" + cfg.Version),
		Body:                 github.String(fmt.Sprintf("Release %s", cfg.Version)),
		Draft:                github.Bool(cfg.GitHub.Release.Draft),
		Prerelease:           github.Bool(cfg.GitHub.Release.Prerelease),
		GenerateReleaseNotes: github.Bool(cfg.GitHub.Release.GenerateNotes),
	}

	rel, _, err := c.gh.Repositories.CreateRelease(ctx, cfg.GitHub.Owner, cfg.GitHub.Repo, release)
	if err != nil {
		return nil, fmt.Errorf("failed to create release: %w", err)
	}

	// Upload assets
	for _, asset := range assets {
		if err := c.uploadAsset(ctx, cfg, rel.GetID(), asset); err != nil {
			return nil, fmt.Errorf("failed to upload asset %s: %w", asset, err)
		}
	}

	return rel, nil
}

func (c *Client) uploadAsset(ctx context.Context, cfg *config.Config, releaseID int64, assetPath string) error {
	file, err := os.Open(assetPath)
	if err != nil {
		return err
	}
	defer file.Close()

	opts := &github.UploadOptions{
		Name: filepath.Base(assetPath),
	}

	_, _, err = c.gh.Repositories.UploadReleaseAsset(ctx, cfg.GitHub.Owner, cfg.GitHub.Repo, releaseID, opts, file)
	return err
}

func (c *Client) UpdateTap(ctx context.Context, cfg *config.Config, formula string) error {
	if !cfg.GitHub.Tap.Enabled {
		return nil
	}

	tapRepo := cfg.GitHub.Tap.Repo
	if tapRepo == "" {
		tapRepo = fmt.Sprintf("%s/homebrew-tap", cfg.GitHub.Owner)
	}

	parts := strings.Split(tapRepo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid tap repo format: %s", tapRepo)
	}
	tapOwner, tapRepoName := parts[0], parts[1]

	// Create repository if it doesn't exist and auto_create is enabled
	if cfg.GitHub.Tap.AutoCreate {
		if err := c.ensureRepository(ctx, tapOwner, tapRepoName, "Homebrew tap for "+cfg.Name); err != nil {
			return fmt.Errorf("failed to ensure tap repository: %w", err)
		}
	}

	// Update formula file
	formulaPath := fmt.Sprintf("Formula/%s.rb", cfg.Name)
	commitMessage := fmt.Sprintf("Update %s to v%s", cfg.Name, cfg.Version)
	
	if cfg.GitHub.Tap.AutoCommit {
		return c.updateFile(ctx, tapOwner, tapRepoName, formulaPath, formula, commitMessage)
	}

	fmt.Printf("✅ Would update tap %s with formula (auto_commit disabled)\n", tapRepo)
	return nil
}

func (c *Client) UpdateBucket(ctx context.Context, cfg *config.Config, manifest string) error {
	if !cfg.GitHub.Bucket.Enabled {
		return nil
	}

	bucketRepo := cfg.GitHub.Bucket.Repo
	if bucketRepo == "" {
		bucketRepo = fmt.Sprintf("%s/scoop-bucket", cfg.GitHub.Owner)
	}

	parts := strings.Split(bucketRepo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid bucket repo format: %s", bucketRepo)
	}
	bucketOwner, bucketRepoName := parts[0], parts[1]

	// Create repository if it doesn't exist and auto_create is enabled
	if cfg.GitHub.Bucket.AutoCreate {
		if err := c.ensureRepository(ctx, bucketOwner, bucketRepoName, "Scoop bucket for "+cfg.Name); err != nil {
			return fmt.Errorf("failed to ensure bucket repository: %w", err)
		}
	}

	// Update manifest file
	manifestPath := fmt.Sprintf("bucket/%s.json", cfg.Name)
	commitMessage := fmt.Sprintf("Update %s to v%s", cfg.Name, cfg.Version)
	
	if cfg.GitHub.Bucket.AutoCommit {
		return c.updateFile(ctx, bucketOwner, bucketRepoName, manifestPath, manifest, commitMessage)
	}

	fmt.Printf("✅ Would update bucket %s with manifest (auto_commit disabled)\n", bucketRepo)
	return nil
}

func (c *Client) ensureRepository(ctx context.Context, owner, repo, description string) error {
	// Check if repository exists
	_, _, err := c.gh.Repositories.Get(ctx, owner, repo)
	if err == nil {
		return nil // Repository exists
	}

	// Create repository
	repository := &github.Repository{
		Name:        github.String(repo),
		Description: github.String(description),
		Private:     github.Bool(false),
	}

	_, _, err = c.gh.Repositories.Create(ctx, "", repository)
	if err != nil {
		return fmt.Errorf("failed to create repository %s/%s: %w", owner, repo, err)
	}

	fmt.Printf("✅ Created repository %s/%s\n", owner, repo)
	return nil
}

func (c *Client) updateFile(ctx context.Context, owner, repo, path, content, commitMessage string) error {
	// Get current file (if exists)
	var currentSHA *string
	fileContent, _, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err == nil && fileContent != nil {
		currentSHA = fileContent.SHA
	}

	// Update or create file
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(commitMessage),
		Content: []byte(content),
		SHA:     currentSHA,
	}

	_, _, err = c.gh.Repositories.CreateFile(ctx, owner, repo, path, opts)
	if err != nil {
		return fmt.Errorf("failed to update file %s: %w", path, err)
	}

	fmt.Printf("✅ Updated %s/%s:%s\n", owner, repo, path)
	return nil
}

func (c *Client) SubmitWingetPR(ctx context.Context, cfg *config.Config, manifests map[string]string) error {
	if !cfg.GitHub.Winget.Enabled || !cfg.GitHub.Winget.AutoPR {
		return nil
	}

	upstreamOwner := "microsoft"
	upstreamRepo := "winget-pkgs"
	forkRepo := cfg.GitHub.Winget.ForkRepo
	
	if forkRepo == "" {
		forkRepo = fmt.Sprintf("%s/winget-pkgs", cfg.GitHub.Owner)
	}

	parts := strings.Split(forkRepo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid fork repo format: %s", forkRepo)
	}
	forkOwner, forkRepoName := parts[0], parts[1]

	// Ensure fork exists
	if err := c.ensureFork(ctx, upstreamOwner, upstreamRepo, forkOwner); err != nil {
		return fmt.Errorf("failed to ensure fork: %w", err)
	}

	// Create branch
	branchName := fmt.Sprintf("%s-%s", strings.ToLower(cfg.Name), cfg.Version)
	if err := c.createBranch(ctx, forkOwner, forkRepoName, branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// Update manifest files
	manifestDir := fmt.Sprintf("manifests/%s/%s/%s/%s", 
		strings.ToLower(string(cfg.Packages.Winget.Publisher[0])),
		cfg.Packages.Winget.Publisher,
		cfg.Packages.Winget.PackageIdentifier,
		cfg.Version)

	for filename, content := range manifests {
		manifestPath := fmt.Sprintf("%s/%s", manifestDir, filename)
		commitMessage := fmt.Sprintf("Add %s version %s", cfg.Packages.Winget.PackageIdentifier, cfg.Version)
		
		if err := c.updateFileOnBranch(ctx, forkOwner, forkRepoName, branchName, manifestPath, content, commitMessage); err != nil {
			return fmt.Errorf("failed to update manifest %s: %w", filename, err)
		}
	}

	// Create pull request
	prTitle := fmt.Sprintf("New version: %s version %s", cfg.Packages.Winget.PackageIdentifier, cfg.Version)
	prBody := fmt.Sprintf(`This PR adds %s version %s to the Windows Package Manager Community Repository.

**Package Information:**
- Package Identifier: %s
- Version: %s
- Publisher: %s

**Validation:**
- [ ] Manifests validated with winget validate
- [ ] Package installs successfully
- [ ] Package uninstalls cleanly

---
*This PR was automatically generated by bagboy*`, 
		cfg.Name, cfg.Version,
		cfg.Packages.Winget.PackageIdentifier, cfg.Version, cfg.Packages.Winget.Publisher)

	pr := &github.NewPullRequest{
		Title: github.String(prTitle),
		Head:  github.String(fmt.Sprintf("%s:%s", forkOwner, branchName)),
		Base:  github.String("master"),
		Body:  github.String(prBody),
	}

	createdPR, _, err := c.gh.PullRequests.Create(ctx, upstreamOwner, upstreamRepo, pr)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	fmt.Printf("✅ Created Winget PR: %s\n", createdPR.GetHTMLURL())
	return nil
}

func (c *Client) ensureFork(ctx context.Context, upstreamOwner, upstreamRepo, forkOwner string) error {
	// Check if fork exists
	_, _, err := c.gh.Repositories.Get(ctx, forkOwner, upstreamRepo)
	if err == nil {
		return nil // Fork exists
	}

	// Create fork
	opts := &github.RepositoryCreateForkOptions{}
	_, _, err = c.gh.Repositories.CreateFork(ctx, upstreamOwner, upstreamRepo, opts)
	if err != nil {
		return fmt.Errorf("failed to create fork: %w", err)
	}

	fmt.Printf("✅ Created fork %s/%s\n", forkOwner, upstreamRepo)
	return nil
}

func (c *Client) createBranch(ctx context.Context, owner, repo, branchName string) error {
	// Get default branch SHA
	repository, _, err := c.gh.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return err
	}

	defaultBranch := repository.GetDefaultBranch()
	ref, _, err := c.gh.Git.GetRef(ctx, owner, repo, fmt.Sprintf("heads/%s", defaultBranch))
	if err != nil {
		return err
	}

	// Create new branch
	newRef := &github.Reference{
		Ref: github.String(fmt.Sprintf("refs/heads/%s", branchName)),
		Object: &github.GitObject{
			SHA: ref.Object.SHA,
		},
	}

	_, _, err = c.gh.Git.CreateRef(ctx, owner, repo, newRef)
	if err != nil {
		// Branch might already exist, which is fine
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	return nil
}

func (c *Client) updateFileOnBranch(ctx context.Context, owner, repo, branch, path, content, commitMessage string) error {
	// Get current file (if exists)
	var currentSHA *string
	fileContent, _, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{
		Ref: branch,
	})
	if err == nil && fileContent != nil {
		currentSHA = fileContent.SHA
	}

	// Update or create file
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(commitMessage),
		Content: []byte(content),
		SHA:     currentSHA,
		Branch:  github.String(branch),
	}

	_, _, err = c.gh.Repositories.CreateFile(ctx, owner, repo, path, opts)
	if err != nil {
		return fmt.Errorf("failed to update file %s: %w", path, err)
	}

	return nil
}
