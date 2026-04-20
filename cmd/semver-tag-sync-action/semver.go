package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// semverRegex matches semantic versioning tags like v1.2.3, v1.2.3-beta, v1.2.3+build.
var semverRegex = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)([-+].*)?$`)

// SemVer represents a parsed semantic version.
type SemVer struct {
	Major        string
	Minor        string
	Patch        string
	Suffix       string // Prerelease and/or build metadata suffix (e.g., "-beta+build")
	Full         string
	IsPrerelease bool // True only if suffix starts with "-" (not for build metadata only)
}

// ParseSemVer parses a semantic version tag and returns its components.
func ParseSemVer(tag string) (*SemVer, error) {
	matches := semverRegex.FindStringSubmatch(tag)
	if matches == nil {
		return nil, fmt.Errorf("tag %q does not match semantic versioning format (expected vX.Y.Z)", tag)
	}
	suffix := ""
	if len(matches) > 4 {
		suffix = matches[4]
	}
	// Per semver spec: prerelease versions have a hyphen suffix (e.g., -beta, -rc.1)
	// Build metadata uses + suffix (e.g., +build.123) and is NOT a prerelease
	isPrerelease := strings.HasPrefix(suffix, "-")
	return &SemVer{
		Major:        matches[1],
		Minor:        matches[2],
		Patch:        matches[3],
		Suffix:       suffix,
		Full:         tag,
		IsPrerelease: isPrerelease,
	}, nil
}

// MajorTag returns the major version tag (e.g., "v1").
func (s *SemVer) MajorTag() string {
	return fmt.Sprintf("v%s", s.Major)
}

// MinorTag returns the minor version tag (e.g., "v1.2").
func (s *SemVer) MinorTag() string {
	return fmt.Sprintf("v%s.%s", s.Major, s.Minor)
}

// SemVerGreaterThan returns true if a represents a higher version than b.
func SemVerGreaterThan(a, b *SemVer) bool {
	aMaj, _ := strconv.Atoi(a.Major)
	bMaj, _ := strconv.Atoi(b.Major)
	if aMaj != bMaj {
		return aMaj > bMaj
	}
	aMin, _ := strconv.Atoi(a.Minor)
	bMin, _ := strconv.Atoi(b.Minor)
	if aMin != bMin {
		return aMin > bMin
	}
	aPat, _ := strconv.Atoi(a.Patch)
	bPat, _ := strconv.Atoi(b.Patch)
	if aPat != bPat {
		return aPat > bPat
	}
	if a.IsPrerelease != b.IsPrerelease {
		return !a.IsPrerelease
	}
	return false
}
