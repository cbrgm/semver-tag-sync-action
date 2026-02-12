package main

import (
	"testing"
)

func TestGetEnvOrDefault(t *testing.T) {
	// Set up test environment variable
	t.Setenv("TEST_ENV_VAR", "env-value")

	tests := []struct {
		name      string
		flagValue string
		envVar    string
		want      string
	}{
		{
			name:      "flag value takes precedence",
			flagValue: "flag-value",
			envVar:    "TEST_ENV_VAR",
			want:      "flag-value",
		},
		{
			name:      "env var used when flag is empty",
			flagValue: "",
			envVar:    "TEST_ENV_VAR",
			want:      "env-value",
		},
		{
			name:      "empty when both are empty",
			flagValue: "",
			envVar:    "NONEXISTENT_VAR",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEnvOrDefault(tt.flagValue, tt.envVar)
			if got != tt.want {
				t.Errorf("getEnvOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				GitHubToken: "token",
				GitHubRepo:  "owner/repo",
				GitRef:      "refs/tags/v1.2.3",
				CommitSHA:   "abc123",
				SyncMajor:   true,
				SyncMinor:   true,
			},
			wantErr: false,
		},
		{
			name: "missing token",
			config: Config{
				GitHubRepo: "owner/repo",
				GitRef:     "refs/tags/v1.2.3",
				CommitSHA:  "abc123",
				SyncMajor:  true,
				SyncMinor:  true,
			},
			wantErr: true,
		},
		{
			name: "missing repo",
			config: Config{
				GitHubToken: "token",
				GitRef:      "refs/tags/v1.2.3",
				CommitSHA:   "abc123",
				SyncMajor:   true,
				SyncMinor:   true,
			},
			wantErr: true,
		},
		{
			name: "missing ref",
			config: Config{
				GitHubToken: "token",
				GitHubRepo:  "owner/repo",
				CommitSHA:   "abc123",
				SyncMajor:   true,
				SyncMinor:   true,
			},
			wantErr: true,
		},
		{
			name: "missing sha",
			config: Config{
				GitHubToken: "token",
				GitHubRepo:  "owner/repo",
				GitRef:      "refs/tags/v1.2.3",
				SyncMajor:   true,
				SyncMinor:   true,
			},
			wantErr: true,
		},
		{
			name: "both sync disabled",
			config: Config{
				GitHubToken: "token",
				GitHubRepo:  "owner/repo",
				GitRef:      "refs/tags/v1.2.3",
				CommitSHA:   "abc123",
				SyncMajor:   false,
				SyncMinor:   false,
			},
			wantErr: true,
		},
		{
			name: "only major sync enabled",
			config: Config{
				GitHubToken: "token",
				GitHubRepo:  "owner/repo",
				GitRef:      "refs/tags/v1.2.3",
				CommitSHA:   "abc123",
				SyncMajor:   true,
				SyncMinor:   false,
			},
			wantErr: false,
		},
		{
			name: "only minor sync enabled",
			config: Config{
				GitHubToken: "token",
				GitHubRepo:  "owner/repo",
				GitRef:      "refs/tags/v1.2.3",
				CommitSHA:   "abc123",
				SyncMajor:   false,
				SyncMinor:   true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
