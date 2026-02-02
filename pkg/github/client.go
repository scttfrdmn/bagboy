package github

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
	// This would implement tap management
	// For now, just a placeholder
	fmt.Printf("Would update tap %s with formula\n", cfg.GitHub.Tap.Repo)
	return nil
}

func (c *Client) UpdateBucket(ctx context.Context, cfg *config.Config, manifest string) error {
	// This would implement bucket management
	// For now, just a placeholder
	fmt.Printf("Would update bucket %s with manifest\n", cfg.GitHub.Bucket.Repo)
	return nil
}
