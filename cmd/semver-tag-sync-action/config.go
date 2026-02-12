package main

import (
	"fmt"
	"os"
)

// Config holds the action configuration.
type Config struct {
	GitHubToken         string
	GitHubRepo          string
	GitRef              string
	CommitSHA           string
	SyncMajor           bool
	SyncMinor           bool
	SkipPrereleases     bool
	DryRun              bool
	GitHubEnterpriseURL string
	LogLevel            string
}

// Validate checks the configuration for required values.
func (c *Config) Validate() error {
	if c.GitHubToken == "" {
		return fmt.Errorf("github token is required (set --github-token or GITHUB_TOKEN)")
	}
	if c.GitHubRepo == "" {
		return fmt.Errorf("github repo is required (set --github-repo or GITHUB_REPOSITORY)")
	}
	if c.GitRef == "" {
		return fmt.Errorf("git ref is required (set --git-ref or GITHUB_REF)")
	}
	if c.CommitSHA == "" {
		return fmt.Errorf("commit sha is required (set --commit-sha or GITHUB_SHA)")
	}
	if !c.SyncMajor && !c.SyncMinor {
		return fmt.Errorf("at least one of --sync-major or --sync-minor must be enabled")
	}
	return nil
}

// getEnvOrDefault returns the flag value if set, otherwise falls back to the environment variable.
func getEnvOrDefault(flagValue, envVar string) string {
	if flagValue != "" {
		return flagValue
	}
	return os.Getenv(envVar)
}
