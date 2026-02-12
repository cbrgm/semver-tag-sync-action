package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	Version   string
	Revision  string
	BuildDate string
	GoVersion = runtime.Version()
	StartTime = time.Now()
)

func main() {
	var (
		githubToken         string
		githubRepo          string
		gitRef              string
		commitSHA           string
		syncMajor           bool
		syncMinor           bool
		skipPrereleases     bool
		dryRun              bool
		githubEnterpriseURL string
		logLevel            string
		showVersion         bool
	)

	flag.StringVar(&githubToken, "github-token", "", "GitHub token for authentication (or set GITHUB_TOKEN)")
	flag.StringVar(&githubRepo, "github-repo", "", "Target repository in owner/repo format (default: GITHUB_REPOSITORY)")
	flag.StringVar(&gitRef, "git-ref", "", "Git reference, e.g., refs/tags/v1.2.3 (default: GITHUB_REF)")
	flag.StringVar(&commitSHA, "commit-sha", "", "Commit SHA to point the tags to (default: GITHUB_SHA)")
	flag.BoolVar(&syncMajor, "sync-major", true, "Sync major version tag (e.g., v1)")
	flag.BoolVar(&syncMinor, "sync-minor", true, "Sync minor version tag (e.g., v1.2)")
	flag.BoolVar(&skipPrereleases, "skip-prereleases", true, "Skip syncing for prerelease versions (e.g., v1.2.3-beta)")
	flag.BoolVar(&dryRun, "dry-run", false, "Perform a dry run without making changes")
	flag.StringVar(&githubEnterpriseURL, "github-enterprise-url", "", "GitHub Enterprise URL (optional)")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.BoolVar(&showVersion, "version", false, "Show version information")

	flag.Parse()

	if showVersion {
		fmt.Printf("semver-tag-sync-action\nVersion: %s %s\nBuildDate: %s\n%s\n", Revision, Version, BuildDate, GoVersion)
		os.Exit(0)
	}

	// Setup logger
	log := setupLogger(logLevel)

	log.Debug("Starting with configuration",
		slog.String("version", Version),
		slog.String("revision", Revision),
		slog.String("build_date", BuildDate),
		slog.String("go_version", GoVersion),
		slog.String("log_level", logLevel),
	)

	// Auto-discover from GitHub Actions environment if not explicitly set
	githubToken = getEnvOrDefault(githubToken, "GITHUB_TOKEN")
	githubRepo = getEnvOrDefault(githubRepo, "GITHUB_REPOSITORY")
	gitRef = getEnvOrDefault(gitRef, "GITHUB_REF")
	commitSHA = getEnvOrDefault(commitSHA, "GITHUB_SHA")

	config := Config{
		GitHubToken:         githubToken,
		GitHubRepo:          githubRepo,
		GitRef:              gitRef,
		CommitSHA:           commitSHA,
		SyncMajor:           syncMajor,
		SyncMinor:           syncMinor,
		SkipPrereleases:     skipPrereleases,
		DryRun:              dryRun,
		GitHubEnterpriseURL: githubEnterpriseURL,
		LogLevel:            logLevel,
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Error("Configuration validation failed",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	// Create GitHub client
	client, err := NewGitHubClient(config.GitHubToken, config.GitHubEnterpriseURL)
	if err != nil {
		log.Error("Failed to create GitHub client",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	action := NewAction(client, config, log)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := action.Run(ctx); err != nil {
		log.Error("Action failed",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
}

// setupLogger creates a new slog.Logger with the specified log level.
func setupLogger(level string) *slog.Logger {
	logLevel := stringToLogLevel(level)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	return slog.New(handler)
}

// stringToLogLevel converts a string to a slog.Level.
func stringToLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
