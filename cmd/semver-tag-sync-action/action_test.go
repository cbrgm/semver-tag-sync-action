package main

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-github/v84/github"
)

// mockGitHubClient is a mock implementation of GitHubClient for testing.
type mockGitHubClient struct {
	getRefFunc    func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error)
	createRefFunc func(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error)
	updateRefFunc func(ctx context.Context, owner, repo, ref string, updateRef github.UpdateRef) (*github.Reference, *github.Response, error)
	listTagsFunc  func(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error)
}

func (m *mockGitHubClient) GetRef(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
	if m.getRefFunc != nil {
		return m.getRefFunc(ctx, owner, repo, ref)
	}
	return nil, &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
}

func (m *mockGitHubClient) CreateRef(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error) {
	if m.createRefFunc != nil {
		return m.createRefFunc(ctx, owner, repo, ref)
	}
	return &github.Reference{}, &github.Response{Response: &http.Response{StatusCode: http.StatusCreated}}, nil
}

func (m *mockGitHubClient) UpdateRef(ctx context.Context, owner, repo, ref string, updateRef github.UpdateRef) (*github.Reference, *github.Response, error) {
	if m.updateRefFunc != nil {
		return m.updateRefFunc(ctx, owner, repo, ref, updateRef)
	}
	return &github.Reference{}, &github.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
}

func (m *mockGitHubClient) ListTags(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
	if m.listTagsFunc != nil {
		return m.listTagsFunc(ctx, owner, repo, opts)
	}
	return nil, &github.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
}

func TestActionRun_CreateNewTags(t *testing.T) {
	var createdRefs []string
	mock := &mockGitHubClient{
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			return nil, &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
		},
		createRefFunc: func(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error) {
			createdRefs = append(createdRefs, ref.Ref)
			return &github.Reference{}, &github.Response{Response: &http.Response{StatusCode: http.StatusCreated}}, nil
		},
	}

	config := Config{
		GitHubRepo: "owner/repo",
		GitRef:     "refs/tags/v1.2.3",
		CommitSHA:  "abc123",
		SyncMajor:  true,
		SyncMinor:  true,
		DryRun:     false,
	}

	action := NewAction(mock, config, nil)

	err := action.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(createdRefs) != 2 {
		t.Fatalf("expected 2 refs to be created, got %d", len(createdRefs))
	}

	expectedRefs := map[string]bool{
		"refs/tags/v1":   false,
		"refs/tags/v1.2": false,
	}
	for _, ref := range createdRefs {
		if _, ok := expectedRefs[ref]; !ok {
			t.Errorf("unexpected ref created: %s", ref)
		}
		expectedRefs[ref] = true
	}
	for ref, created := range expectedRefs {
		if !created {
			t.Errorf("expected ref %s to be created", ref)
		}
	}
}

func TestActionRun_UpdateExistingTags(t *testing.T) {
	var updatedRefs []string
	mock := &mockGitHubClient{
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			return &github.Reference{}, &github.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
		},
		updateRefFunc: func(ctx context.Context, owner, repo, ref string, updateRef github.UpdateRef) (*github.Reference, *github.Response, error) {
			if updateRef.Force == nil || !*updateRef.Force {
				t.Error("expected force=true for update")
			}
			updatedRefs = append(updatedRefs, ref)
			return &github.Reference{}, &github.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
		},
	}

	config := Config{
		GitHubRepo: "owner/repo",
		GitRef:     "refs/tags/v1.2.3",
		CommitSHA:  "abc123",
		SyncMajor:  true,
		SyncMinor:  true,
		DryRun:     false,
	}

	action := NewAction(mock, config, nil)

	err := action.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(updatedRefs) != 2 {
		t.Fatalf("expected 2 refs to be updated, got %d", len(updatedRefs))
	}
}

func TestActionRun_DryRun(t *testing.T) {
	mock := &mockGitHubClient{
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			return nil, &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
		},
		createRefFunc: func(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error) {
			t.Error("createRef should not be called in dry-run mode")
			return nil, nil, nil
		},
		updateRefFunc: func(ctx context.Context, owner, repo, ref string, updateRef github.UpdateRef) (*github.Reference, *github.Response, error) {
			t.Error("updateRef should not be called in dry-run mode")
			return nil, nil, nil
		},
	}

	config := Config{
		GitHubRepo: "owner/repo",
		GitRef:     "refs/tags/v1.2.3",
		CommitSHA:  "abc123",
		SyncMajor:  true,
		SyncMinor:  true,
		DryRun:     true,
	}

	action := NewAction(mock, config, nil)

	err := action.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestActionRun_SyncMajorOnly(t *testing.T) {
	var createdRefs []string
	mock := &mockGitHubClient{
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			return nil, &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
		},
		createRefFunc: func(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error) {
			createdRefs = append(createdRefs, ref.Ref)
			return &github.Reference{}, &github.Response{Response: &http.Response{StatusCode: http.StatusCreated}}, nil
		},
	}

	config := Config{
		GitHubRepo: "owner/repo",
		GitRef:     "refs/tags/v1.2.3",
		CommitSHA:  "abc123",
		SyncMajor:  true,
		SyncMinor:  false,
		DryRun:     false,
	}

	action := NewAction(mock, config, nil)

	err := action.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(createdRefs) != 1 {
		t.Fatalf("expected 1 ref to be created, got %d", len(createdRefs))
	}

	if createdRefs[0] != "refs/tags/v1" {
		t.Errorf("expected refs/tags/v1, got %s", createdRefs[0])
	}
}

func TestActionRun_SyncMinorOnly(t *testing.T) {
	var createdRefs []string
	mock := &mockGitHubClient{
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			return nil, &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
		},
		createRefFunc: func(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error) {
			createdRefs = append(createdRefs, ref.Ref)
			return &github.Reference{}, &github.Response{Response: &http.Response{StatusCode: http.StatusCreated}}, nil
		},
	}

	config := Config{
		GitHubRepo: "owner/repo",
		GitRef:     "refs/tags/v1.2.3",
		CommitSHA:  "abc123",
		SyncMajor:  false,
		SyncMinor:  true,
		DryRun:     false,
	}

	action := NewAction(mock, config, nil)

	err := action.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(createdRefs) != 1 {
		t.Fatalf("expected 1 ref to be created, got %d", len(createdRefs))
	}

	if createdRefs[0] != "refs/tags/v1.2" {
		t.Errorf("expected refs/tags/v1.2, got %s", createdRefs[0])
	}
}

func TestActionRun_InvalidRef(t *testing.T) {
	mock := &mockGitHubClient{}

	config := Config{
		GitHubRepo: "owner/repo",
		GitRef:     "refs/heads/main",
		CommitSHA:  "abc123",
		SyncMajor:  true,
		SyncMinor:  true,
	}

	action := NewAction(mock, config, nil)

	err := action.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for non-tag ref")
	}
}

func TestActionRun_InvalidSemVer(t *testing.T) {
	mock := &mockGitHubClient{}

	config := Config{
		GitHubRepo: "owner/repo",
		GitRef:     "refs/tags/v1.2",
		CommitSHA:  "abc123",
		SyncMajor:  true,
		SyncMinor:  true,
	}

	action := NewAction(mock, config, nil)

	err := action.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid semver tag")
	}
}

func TestActionRun_SkipPrereleases(t *testing.T) {
	mock := &mockGitHubClient{
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			t.Error("getRef should not be called when skipping prereleases")
			return nil, nil, nil
		},
		createRefFunc: func(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error) {
			t.Error("createRef should not be called when skipping prereleases")
			return nil, nil, nil
		},
	}

	config := Config{
		GitHubRepo:      "owner/repo",
		GitRef:          "refs/tags/v1.2.3-beta",
		CommitSHA:       "abc123",
		SyncMajor:       true,
		SyncMinor:       true,
		SkipPrereleases: true,
		DryRun:          false,
	}

	action := NewAction(mock, config, nil)

	err := action.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestActionRun_ProcessPrereleases(t *testing.T) {
	var createdRefs []string
	mock := &mockGitHubClient{
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			return nil, &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
		},
		createRefFunc: func(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error) {
			createdRefs = append(createdRefs, ref.Ref)
			return &github.Reference{}, &github.Response{Response: &http.Response{StatusCode: http.StatusCreated}}, nil
		},
	}

	config := Config{
		GitHubRepo:      "owner/repo",
		GitRef:          "refs/tags/v1.2.3-beta",
		CommitSHA:       "abc123",
		SyncMajor:       true,
		SyncMinor:       true,
		SkipPrereleases: false, // Process prereleases
		DryRun:          false,
	}

	action := NewAction(mock, config, nil)

	err := action.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(createdRefs) != 2 {
		t.Fatalf("expected 2 refs to be created, got %d", len(createdRefs))
	}
}

func TestActionRun_APIError(t *testing.T) {
	mock := &mockGitHubClient{
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			return &github.Reference{}, &github.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
		},
		updateRefFunc: func(ctx context.Context, owner, repo, ref string, updateRef github.UpdateRef) (*github.Reference, *github.Response, error) {
			return nil, &github.Response{Response: &http.Response{StatusCode: http.StatusForbidden}}, errors.New("forbidden")
		},
	}

	config := Config{
		GitHubRepo: "owner/repo",
		GitRef:     "refs/tags/v1.2.3",
		CommitSHA:  "abc123",
		SyncMajor:  true,
		SyncMinor:  true,
		DryRun:     false,
	}

	action := NewAction(mock, config, nil)

	err := action.Run(context.Background())
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

func TestActionRun_NetworkError(t *testing.T) {
	mock := &mockGitHubClient{
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			// Simulate network error - nil response
			return nil, nil, errors.New("network error")
		},
	}

	config := Config{
		GitHubRepo: "owner/repo",
		GitRef:     "refs/tags/v1.2.3",
		CommitSHA:  "abc123",
		SyncMajor:  true,
		SyncMinor:  true,
		DryRun:     false,
	}

	action := NewAction(mock, config, nil)

	err := action.Run(context.Background())
	if err == nil {
		t.Fatal("expected error from network failure")
	}
}

func makeTag(name, sha string) *github.RepositoryTag {
	return &github.RepositoryTag{
		Name: github.Ptr(name),
		Commit: &github.Commit{
			SHA: github.Ptr(sha),
		},
	}
}

func TestActionRunAll_CreatesAllMajorMinorTags(t *testing.T) {
	var createdRefs []string
	mock := &mockGitHubClient{
		listTagsFunc: func(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
			tags := []*github.RepositoryTag{
				makeTag("v1.0.0", "sha100"),
				makeTag("v1.0.1", "sha101"),
				makeTag("v1.1.0", "sha110"),
				makeTag("v2.0.0", "sha200"),
				makeTag("not-semver", "shaxxx"),
			}
			return tags, &github.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
		},
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			return nil, &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
		},
		createRefFunc: func(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error) {
			createdRefs = append(createdRefs, ref.Ref+"="+ref.SHA)
			return &github.Reference{}, &github.Response{Response: &http.Response{StatusCode: http.StatusCreated}}, nil
		},
	}

	config := Config{
		GitHubRepo:  "owner/repo",
		SyncMajor:   true,
		SyncMinor:   true,
		SyncAllTags: true,
	}

	action := NewAction(mock, config, nil)
	err := action.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	expectedMajor := map[string]string{
		"refs/tags/v1": "sha110",
		"refs/tags/v2": "sha200",
	}
	expectedMinor := map[string]string{
		"refs/tags/v1.0": "sha101",
		"refs/tags/v1.1": "sha110",
		"refs/tags/v2.0": "sha200",
	}

	all := make(map[string]string)
	for k, v := range expectedMajor {
		all[k] = v
	}
	for k, v := range expectedMinor {
		all[k] = v
	}

	if len(createdRefs) != len(all) {
		t.Fatalf("expected %d refs to be created, got %d: %v", len(all), len(createdRefs), createdRefs)
	}

	for _, entry := range createdRefs {
		found := false
		for ref, sha := range all {
			if entry == ref+"="+sha {
				found = true
				delete(all, ref)
				break
			}
		}
		if !found {
			t.Errorf("unexpected ref created: %s", entry)
		}
	}
}

func TestActionRunAll_SkipsTagsAlreadyPointingToCorrectSHA(t *testing.T) {
	mock := &mockGitHubClient{
		listTagsFunc: func(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
			return []*github.RepositoryTag{
				makeTag("v1.0.0", "sha100"),
			}, &github.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
		},
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			return &github.Reference{
				Object: &github.GitObject{SHA: github.Ptr("sha100")},
			}, &github.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
		},
		createRefFunc: func(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error) {
			t.Error("createRef should not be called when tag already points to correct SHA")
			return nil, nil, nil
		},
		updateRefFunc: func(ctx context.Context, owner, repo, ref string, updateRef github.UpdateRef) (*github.Reference, *github.Response, error) {
			t.Error("updateRef should not be called when tag already points to correct SHA")
			return nil, nil, nil
		},
	}

	config := Config{
		GitHubRepo:  "owner/repo",
		SyncMajor:   true,
		SyncMinor:   true,
		SyncAllTags: true,
	}

	action := NewAction(mock, config, nil)
	err := action.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestActionRunAll_SkipsPrereleases(t *testing.T) {
	var createdRefs []string
	mock := &mockGitHubClient{
		listTagsFunc: func(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
			return []*github.RepositoryTag{
				makeTag("v1.0.0", "sha100"),
				makeTag("v1.0.1-beta", "sha101beta"),
			}, &github.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
		},
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			return nil, &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
		},
		createRefFunc: func(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error) {
			createdRefs = append(createdRefs, ref.Ref+"="+ref.SHA)
			return &github.Reference{}, &github.Response{Response: &http.Response{StatusCode: http.StatusCreated}}, nil
		},
	}

	config := Config{
		GitHubRepo:      "owner/repo",
		SyncMajor:       true,
		SyncMinor:       true,
		SyncAllTags:     true,
		SkipPrereleases: true,
	}

	action := NewAction(mock, config, nil)
	err := action.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	for _, entry := range createdRefs {
		if entry == "refs/tags/v1=sha101beta" || entry == "refs/tags/v1.0=sha101beta" {
			t.Errorf("prerelease SHA should not have been used: %s", entry)
		}
	}
}

func TestActionRunAll_Pagination(t *testing.T) {
	var createdRefs []string
	mock := &mockGitHubClient{
		listTagsFunc: func(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
			if opts.Page <= 0 || opts.Page == 1 {
				return []*github.RepositoryTag{
					makeTag("v1.0.0", "sha100"),
				}, &github.Response{
					Response: &http.Response{StatusCode: http.StatusOK},
					NextPage: 2,
				}, nil
			}
			return []*github.RepositoryTag{
				makeTag("v1.0.1", "sha101"),
			}, &github.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
		},
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			return nil, &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
		},
		createRefFunc: func(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error) {
			createdRefs = append(createdRefs, ref.Ref+"="+ref.SHA)
			return &github.Reference{}, &github.Response{Response: &http.Response{StatusCode: http.StatusCreated}}, nil
		},
	}

	config := Config{
		GitHubRepo:  "owner/repo",
		SyncMajor:   true,
		SyncMinor:   true,
		SyncAllTags: true,
	}

	action := NewAction(mock, config, nil)
	err := action.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	foundMajor := false
	foundMinor := false
	for _, entry := range createdRefs {
		if entry == "refs/tags/v1=sha101" {
			foundMajor = true
		}
		if entry == "refs/tags/v1.0=sha101" {
			foundMinor = true
		}
	}
	if !foundMajor {
		t.Error("expected v1 tag pointing to sha101 (latest from pagination)")
	}
	if !foundMinor {
		t.Error("expected v1.0 tag pointing to sha101 (latest from pagination)")
	}
}

func TestActionRunAll_DryRun(t *testing.T) {
	mock := &mockGitHubClient{
		listTagsFunc: func(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
			return []*github.RepositoryTag{
				makeTag("v1.0.0", "sha100"),
			}, &github.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
		},
		getRefFunc: func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
			return nil, &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
		},
		createRefFunc: func(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error) {
			t.Error("createRef should not be called in dry-run mode")
			return nil, nil, nil
		},
		updateRefFunc: func(ctx context.Context, owner, repo, ref string, updateRef github.UpdateRef) (*github.Reference, *github.Response, error) {
			t.Error("updateRef should not be called in dry-run mode")
			return nil, nil, nil
		},
	}

	config := Config{
		GitHubRepo:  "owner/repo",
		SyncMajor:   true,
		SyncMinor:   true,
		SyncAllTags: true,
		DryRun:      true,
	}

	action := NewAction(mock, config, nil)
	err := action.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}
