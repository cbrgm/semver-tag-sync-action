package main

import (
	"testing"
)

func TestExtractTagFromRef(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		want    string
		wantErr bool
	}{
		{
			name:    "valid tag ref",
			ref:     "refs/tags/v1.2.3",
			want:    "v1.2.3",
			wantErr: false,
		},
		{
			name:    "valid tag ref with prerelease",
			ref:     "refs/tags/v1.2.3-beta",
			want:    "v1.2.3-beta",
			wantErr: false,
		},
		{
			name:    "branch ref",
			ref:     "refs/heads/main",
			want:    "",
			wantErr: true,
		},
		{
			name:    "pull request ref",
			ref:     "refs/pull/123/merge",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty ref",
			ref:     "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractTagFromRef(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTagFromRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractTagFromRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRepository(t *testing.T) {
	tests := []struct {
		name      string
		repo      string
		wantOwner string
		wantName  string
		wantErr   bool
	}{
		{
			name:      "valid repository",
			repo:      "owner/repo",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "valid repository with dashes",
			repo:      "my-org/my-repo",
			wantOwner: "my-org",
			wantName:  "my-repo",
			wantErr:   false,
		},
		{
			name:    "missing repo name",
			repo:    "owner/",
			wantErr: true,
		},
		{
			name:    "missing owner",
			repo:    "/repo",
			wantErr: true,
		},
		{
			name:    "no slash",
			repo:    "ownerrepo",
			wantErr: true,
		},
		{
			name:    "too many slashes",
			repo:    "owner/repo/extra",
			wantErr: true,
		},
		{
			name:    "empty string",
			repo:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, name, err := parseRepository(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if owner != tt.wantOwner {
				t.Errorf("parseRepository() owner = %v, want %v", owner, tt.wantOwner)
			}
			if name != tt.wantName {
				t.Errorf("parseRepository() name = %v, want %v", name, tt.wantName)
			}
		})
	}
}
