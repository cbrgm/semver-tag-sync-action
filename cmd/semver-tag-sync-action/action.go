package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/go-github/v85/github"
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
	if a.config.SyncAllTags {
		return a.runAll(ctx)
	}

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
	return a.syncTagToSHA(ctx, owner, repo, tag, a.config.CommitSHA)
}

// syncTagToSHA creates or updates a tag to point to the given commit SHA.
func (a *Action) syncTagToSHA(ctx context.Context, owner, repo, tag, sha string) error {
	refName := fmt.Sprintf("tags/%s", tag)
	fullRefName := fmt.Sprintf("refs/tags/%s", tag)

	a.log.Debug("Checking if tag exists",
		slog.String("tag", tag),
		slog.String("ref_name", refName),
	)

	ref, resp, err := a.client.GetRef(ctx, owner, repo, refName)
	tagExists := err == nil

	if err != nil {
		if resp == nil || resp.StatusCode != http.StatusNotFound {
			a.log.Error("Failed to check if tag exists",
				slog.String("tag", tag),
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("failed to check if tag %s exists: %w", tag, err)
		}
		a.log.Debug("Tag does not exist, will create",
			slog.String("tag", tag),
		)
	} else {
		if ref != nil && ref.Object != nil && ref.Object.GetSHA() == sha {
			a.log.Info("Tag already points to correct SHA, skipping",
				slog.String("tag", tag),
				slog.String("commit_sha", sha),
			)
			return nil
		}
		a.log.Debug("Tag already exists, will update",
			slog.String("tag", tag),
		)
	}

	if a.config.DryRun {
		if tagExists {
			a.log.Info("[dry-run] Would update tag",
				slog.String("tag", tag),
				slog.String("commit_sha", sha),
			)
		} else {
			a.log.Info("[dry-run] Would create tag",
				slog.String("tag", tag),
				slog.String("commit_sha", sha),
			)
		}
		return nil
	}

	if tagExists {
		a.log.Info("Updating tag",
			slog.String("tag", tag),
			slog.String("commit_sha", sha),
		)
		updateRef := github.UpdateRef{
			SHA:   sha,
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
			slog.String("commit_sha", sha),
		)
		createRef := github.CreateRef{
			Ref: fullRefName,
			SHA: sha,
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

// tagWithSHA associates a parsed semver tag with its commit SHA.
type tagWithSHA struct {
	semver *SemVer
	sha    string
}

// runAll syncs major/minor tags for all existing semver tags in the repository.
func (a *Action) runAll(ctx context.Context) error {
	a.log.Info("Starting semver tag sync for all tags",
		slog.String("repo", a.config.GitHubRepo),
		slog.Bool("sync_major", a.config.SyncMajor),
		slog.Bool("sync_minor", a.config.SyncMinor),
		slog.Bool("skip_prereleases", a.config.SkipPrereleases),
		slog.Bool("dry_run", a.config.DryRun),
	)

	owner, repo, err := parseRepository(a.config.GitHubRepo)
	if err != nil {
		return err
	}

	majorLatest, minorLatest, err := a.collectLatestTags(ctx, owner, repo)
	if err != nil {
		return err
	}

	var syncErrors []error
	syncErrors = append(syncErrors, a.syncTagMap(ctx, owner, repo, majorLatest, "major")...)
	syncErrors = append(syncErrors, a.syncTagMap(ctx, owner, repo, minorLatest, "minor")...)

	if len(syncErrors) > 0 {
		return errors.Join(syncErrors...)
	}

	a.log.Info("Semver tag sync for all tags completed successfully")
	return nil
}

// collectLatestTags fetches all tags and returns maps of the latest version per major and minor group.
func (a *Action) collectLatestTags(ctx context.Context, owner, repo string) (majorLatest, minorLatest map[string]*tagWithSHA, err error) {
	majorLatest = make(map[string]*tagWithSHA)
	minorLatest = make(map[string]*tagWithSHA)

	page := 1
	totalTags := 0
	for {
		tags, resp, err := a.client.ListTags(ctx, owner, repo, &github.ListOptions{
			Page:    page,
			PerPage: 100,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list tags (page %d): %w", page, err)
		}

		for _, tag := range tags {
			a.processTag(tag, majorLatest, minorLatest)
		}

		totalTags += len(tags)

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	a.log.Info("Fetched all tags",
		slog.Int("total_tags", totalTags),
		slog.Int("major_groups", len(majorLatest)),
		slog.Int("minor_groups", len(minorLatest)),
	)
	return majorLatest, minorLatest, nil
}

// processTag parses a single repository tag and updates the major/minor latest maps if applicable.
func (a *Action) processTag(tag *github.RepositoryTag, majorLatest, minorLatest map[string]*tagWithSHA) {
	name := tag.GetName()
	sv, err := ParseSemVer(name)
	if err != nil {
		a.log.Debug("Skipping non-semver tag", slog.String("tag", name))
		return
	}

	if sv.IsPrerelease && a.config.SkipPrereleases {
		a.log.Debug("Skipping prerelease tag", slog.String("tag", name))
		return
	}

	sha := tag.GetCommit().GetSHA()
	if sha == "" {
		a.log.Debug("Skipping tag with no commit SHA", slog.String("tag", name))
		return
	}

	entry := &tagWithSHA{semver: sv, sha: sha}

	if a.config.SyncMajor {
		majorKey := sv.MajorTag()
		if existing, ok := majorLatest[majorKey]; !ok || SemVerGreaterThan(sv, existing.semver) {
			majorLatest[majorKey] = entry
		}
	}

	if a.config.SyncMinor {
		minorKey := sv.MinorTag()
		if existing, ok := minorLatest[minorKey]; !ok || SemVerGreaterThan(sv, existing.semver) {
			minorLatest[minorKey] = entry
		}
	}
}

// syncTagMap syncs all tags in the given map, returning any errors encountered.
func (a *Action) syncTagMap(ctx context.Context, owner, repo string, tagMap map[string]*tagWithSHA, label string) []error {
	var errs []error
	for tagName, entry := range tagMap {
		a.log.Debug("Syncing "+label+" tag",
			slog.String("tag", tagName),
			slog.String("from_version", entry.semver.Full),
			slog.String("commit_sha", entry.sha),
		)
		if err := a.syncTagToSHA(ctx, owner, repo, tagName, entry.sha); err != nil {
			a.log.Error("Failed to sync "+label+" tag",
				slog.String("tag", tagName),
				slog.String("error", err.Error()),
			)
			errs = append(errs, fmt.Errorf("failed to sync %s tag %s: %w", label, tagName, err))
		}
	}
	return errs
}
