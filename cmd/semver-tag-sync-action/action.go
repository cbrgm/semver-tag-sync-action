package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/go-github/v83/github"
)

// Action performs the semver tag sync.
type Action struct {
	client GitHubClient
	config Config
	log    *slog.Logger
}

// NewAction creates a new Action instance.
func NewAction(client GitHubClient, config Config, log *slog.Logger) *Action {
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Action{
		client: client,
		config: config,
		log:    log,
	}
}

// Run executes the action.
func (a *Action) Run(ctx context.Context) error {
	a.log.Info("Starting semver tag sync action",
		slog.String("repo", a.config.GitHubRepo),
		slog.String("ref", a.config.GitRef),
		slog.Bool("sync_major", a.config.SyncMajor),
		slog.Bool("sync_minor", a.config.SyncMinor),
		slog.Bool("skip_prereleases", a.config.SkipPrereleases),
		slog.Bool("dry_run", a.config.DryRun),
	)

	// Extract tag from ref
	tag, err := extractTagFromRef(a.config.GitRef)
	if err != nil {
		a.log.Error("Failed to extract tag from ref",
			slog.String("ref", a.config.GitRef),
			slog.String("error", err.Error()),
		)
		return err
	}

	a.log.Debug("Extracted tag from ref",
		slog.String("tag", tag),
		slog.String("ref", a.config.GitRef),
	)

	// Parse semantic version
	semver, err := ParseSemVer(tag)
	if err != nil {
		a.log.Error("Failed to parse semantic version",
			slog.String("tag", tag),
			slog.String("error", err.Error()),
		)
		return err
	}

	a.log.Debug("Parsed semantic version",
		slog.String("tag", semver.Full),
		slog.String("major", semver.Major),
		slog.String("minor", semver.Minor),
		slog.String("patch", semver.Patch),
		slog.Bool("is_prerelease", semver.IsPrerelease),
		slog.String("suffix", semver.Suffix),
	)

	// Skip prereleases if configured
	if semver.IsPrerelease && a.config.SkipPrereleases {
		a.log.Info("Skipping prerelease tag",
			slog.String("tag", semver.Full),
			slog.String("suffix", semver.Suffix),
		)
		return nil
	}

	a.log.Info("Processing tag",
		slog.String("tag", semver.Full),
		slog.String("major", semver.Major),
		slog.String("minor", semver.Minor),
		slog.String("patch", semver.Patch),
	)

	// Parse owner/repo
	owner, repo, err := parseRepository(a.config.GitHubRepo)
	if err != nil {
		a.log.Error("Failed to parse repository",
			slog.String("repo", a.config.GitHubRepo),
			slog.String("error", err.Error()),
		)
		return err
	}

	a.log.Debug("Parsed repository",
		slog.String("owner", owner),
		slog.String("repo", repo),
	)

	var syncErrors []error

	// Sync major version tag
	if a.config.SyncMajor {
		majorTag := semver.MajorTag()
		a.log.Debug("Syncing major version tag",
			slog.String("major_tag", majorTag),
			slog.String("commit_sha", a.config.CommitSHA),
		)
		if err := a.syncTag(ctx, owner, repo, majorTag); err != nil {
			a.log.Error("Failed to sync major tag",
				slog.String("tag", majorTag),
				slog.String("error", err.Error()),
			)
			syncErrors = append(syncErrors, fmt.Errorf("failed to sync major tag %s: %w", majorTag, err))
		}
	}

	// Sync minor version tag
	if a.config.SyncMinor {
		minorTag := semver.MinorTag()
		a.log.Debug("Syncing minor version tag",
			slog.String("minor_tag", minorTag),
			slog.String("commit_sha", a.config.CommitSHA),
		)
		if err := a.syncTag(ctx, owner, repo, minorTag); err != nil {
			a.log.Error("Failed to sync minor tag",
				slog.String("tag", minorTag),
				slog.String("error", err.Error()),
			)
			syncErrors = append(syncErrors, fmt.Errorf("failed to sync minor tag %s: %w", minorTag, err))
		}
	}

	if len(syncErrors) > 0 {
		return errors.Join(syncErrors...)
	}

	a.log.Info("Semver tag sync completed successfully")
	return nil
}

// syncTag creates or updates a tag to point to the configured commit SHA.
func (a *Action) syncTag(ctx context.Context, owner, repo, tag string) error {
	refName := fmt.Sprintf("tags/%s", tag)
	fullRefName := fmt.Sprintf("refs/tags/%s", tag)

	a.log.Debug("Checking if tag exists",
		slog.String("tag", tag),
		slog.String("ref_name", refName),
	)

	// Check if tag exists
	_, resp, err := a.client.GetRef(ctx, owner, repo, refName)
	tagExists := err == nil

	if err != nil {
		// Only treat as "not found" if we got a 404 response
		// Any other error (including nil response) should be reported
		if resp == nil || resp.StatusCode != http.StatusNotFound {
			a.log.Error("Failed to check if tag exists",
				slog.String("tag", tag),
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("failed to check if tag %s exists: %w", tag, err)
		}
		// Tag doesn't exist (404), which is fine - we'll create it
		a.log.Debug("Tag does not exist, will create",
			slog.String("tag", tag),
		)
	} else {
		a.log.Debug("Tag already exists, will update",
			slog.String("tag", tag),
		)
	}

	if a.config.DryRun {
		if tagExists {
			a.log.Info("[dry-run] Would update tag",
				slog.String("tag", tag),
				slog.String("commit_sha", a.config.CommitSHA),
			)
		} else {
			a.log.Info("[dry-run] Would create tag",
				slog.String("tag", tag),
				slog.String("commit_sha", a.config.CommitSHA),
			)
		}
		return nil
	}

	if tagExists {
		a.log.Info("Updating tag",
			slog.String("tag", tag),
			slog.String("commit_sha", a.config.CommitSHA),
		)
		updateRef := github.UpdateRef{
			SHA:   a.config.CommitSHA,
			Force: github.Ptr(true),
		}
		_, _, err = a.client.UpdateRef(ctx, owner, repo, refName, updateRef)
		if err != nil {
			return fmt.Errorf("failed to update tag %s: %w", tag, err)
		}
		a.log.Info("Successfully updated tag",
			slog.String("tag", tag),
		)
	} else {
		a.log.Info("Creating tag",
			slog.String("tag", tag),
			slog.String("commit_sha", a.config.CommitSHA),
		)
		createRef := github.CreateRef{
			Ref: fullRefName,
			SHA: a.config.CommitSHA,
		}
		_, _, err = a.client.CreateRef(ctx, owner, repo, createRef)
		if err != nil {
			return fmt.Errorf("failed to create tag %s: %w", tag, err)
		}
		a.log.Info("Successfully created tag",
			slog.String("tag", tag),
		)
	}

	return nil
}
