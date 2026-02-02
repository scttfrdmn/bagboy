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
