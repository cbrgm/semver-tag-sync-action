package main

import (
	"testing"
)

func TestParseSemVer(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		wantMajor string
		wantMinor string
		wantPatch string
		wantErr   bool
	}{
		{
			name:      "valid semver v1.2.3",
			tag:       "v1.2.3",
			wantMajor: "1",
			wantMinor: "2",
			wantPatch: "3",
			wantErr:   false,
		},
		{
			name:      "valid semver v0.0.1",
			tag:       "v0.0.1",
			wantMajor: "0",
			wantMinor: "0",
			wantPatch: "1",
			wantErr:   false,
		},
		{
			name:      "valid semver v10.20.30",
			tag:       "v10.20.30",
			wantMajor: "10",
			wantMinor: "20",
			wantPatch: "30",
			wantErr:   false,
		},
		{
			name:      "valid semver with prerelease",
			tag:       "v1.2.3-beta",
			wantMajor: "1",
			wantMinor: "2",
			wantPatch: "3",
			wantErr:   false,
		},
		{
			name:      "valid semver with build metadata",
			tag:       "v1.2.3+build.123",
			wantMajor: "1",
			wantMinor: "2",
			wantPatch: "3",
			wantErr:   false,
		},
		{
			name:      "valid semver with prerelease and build",
			tag:       "v1.2.3-alpha.1+build",
			wantMajor: "1",
			wantMinor: "2",
			wantPatch: "3",
			wantErr:   false,
		},
		{
			name:    "invalid - missing v prefix",
			tag:     "1.2.3",
			wantErr: true,
		},
		{
			name:    "invalid - only major.minor",
			tag:     "v1.2",
			wantErr: true,
		},
		{
			name:    "invalid - only major",
			tag:     "v1",
			wantErr: true,
		},
		{
			name:    "invalid - empty string",
			tag:     "",
			wantErr: true,
		},
		{
			name:    "invalid - non-numeric",
			tag:     "v1.2.abc",
			wantErr: true,
		},
		{
			name:      "leading zeros are accepted",
			tag:       "v01.02.03",
			wantMajor: "01",
			wantMinor: "02",
			wantPatch: "03",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			semver, err := ParseSemVer(tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSemVer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if semver.Major != tt.wantMajor {
				t.Errorf("ParseSemVer() Major = %v, want %v", semver.Major, tt.wantMajor)
			}
			if semver.Minor != tt.wantMinor {
				t.Errorf("ParseSemVer() Minor = %v, want %v", semver.Minor, tt.wantMinor)
			}
			if semver.Patch != tt.wantPatch {
				t.Errorf("ParseSemVer() Patch = %v, want %v", semver.Patch, tt.wantPatch)
			}
		})
	}
}

func TestSemVerTags(t *testing.T) {
	semver := &SemVer{
		Major: "1",
		Minor: "2",
		Patch: "3",
		Full:  "v1.2.3",
	}

	if got := semver.MajorTag(); got != "v1" {
		t.Errorf("MajorTag() = %v, want v1", got)
	}

	if got := semver.MinorTag(); got != "v1.2" {
		t.Errorf("MinorTag() = %v, want v1.2", got)
	}
}

func TestParseSemVer_PrereleaseAndBuildMetadata(t *testing.T) {
	tests := []struct {
		name         string
		tag          string
		wantSuffix   string
		wantIsPrerel bool
	}{
		{
			name:         "stable release",
			tag:          "v1.2.3",
			wantSuffix:   "",
			wantIsPrerel: false,
		},
		{
			name:         "beta prerelease",
			tag:          "v1.2.3-beta",
			wantSuffix:   "-beta",
			wantIsPrerel: true,
		},
		{
			name:         "alpha with number",
			tag:          "v1.2.3-alpha.1",
			wantSuffix:   "-alpha.1",
			wantIsPrerel: true,
		},
		{
			name:         "rc prerelease",
			tag:          "v1.2.3-rc.1",
			wantSuffix:   "-rc.1",
			wantIsPrerel: true,
		},
		{
			name:         "build metadata only - NOT a prerelease",
			tag:          "v1.2.3+build.123",
			wantSuffix:   "+build.123",
			wantIsPrerel: false, // Build metadata alone is NOT a prerelease per semver spec
		},
		{
			name:         "prerelease with build metadata",
			tag:          "v1.2.3-beta+build",
			wantSuffix:   "-beta+build",
			wantIsPrerel: true, // Has prerelease suffix (-beta), so IS a prerelease
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			semver, err := ParseSemVer(tt.tag)
			if err != nil {
				t.Fatalf("ParseSemVer() error = %v", err)
			}
			if semver.Suffix != tt.wantSuffix {
				t.Errorf("ParseSemVer() Suffix = %v, want %v", semver.Suffix, tt.wantSuffix)
			}
			if semver.IsPrerelease != tt.wantIsPrerel {
				t.Errorf("ParseSemVer() IsPrerelease = %v, want %v", semver.IsPrerelease, tt.wantIsPrerel)
			}
		})
	}
}
