# semver-tag-sync-action

[![GitHub release](https://img.shields.io/github/release/cbrgm/semver-tag-sync-action.svg)](https://github.com/cbrgm/semver-tag-sync-action)
[![Go Report Card](https://goreportcard.com/badge/github.com/cbrgm/semver-tag-sync-action)](https://goreportcard.com/report/github.com/cbrgm/semver-tag-sync-action)
[![go-lint-test](https://github.com/cbrgm/semver-tag-sync-action/actions/workflows/go-lint-test.yml/badge.svg)](https://github.com/cbrgm/semver-tag-sync-action/actions/workflows/go-lint-test.yml)
[![go-binaries](https://github.com/cbrgm/semver-tag-sync-action/actions/workflows/go-binaries.yml/badge.svg)](https://github.com/cbrgm/semver-tag-sync-action/actions/workflows/go-binaries.yml)
[![container](https://github.com/cbrgm/semver-tag-sync-action/actions/workflows/container.yml/badge.svg)](https://github.com/cbrgm/semver-tag-sync-action/actions/workflows/container.yml)

**Automatically sync major and minor version tags when semantic versioning tags are pushed to your repository.**

This GitHub Action keeps your major version tags (e.g., `v1`) and minor version tags (e.g., `v1.2`) in sync with your semantic version releases (e.g., `v1.2.3`). This is particularly useful for GitHub Action authors who want to provide version-agnostic references to their actions.

## Table of Contents

- [How It Works](#how-it-works)
- [Inputs](#inputs)
- [Workflow Usage](#workflow-usage)
  - [Sync Only Major Version](#sync-only-major-version)
  - [Sync Only Minor Version](#sync-only-minor-version)
  - [Include Prerelease Versions](#include-prerelease-versions)
  - [Dry Run Mode](#dry-run-mode)
  - [Cross-Repository Sync](#cross-repository-sync)
- [Container Usage](#container-usage)
- [Local Development](#local-development)
- [Contributing & License](#contributing--license)

## How It Works

When you push a semantic versioning tag like `v1.2.3`, this action will:

- **Sync Major Tag**: Create or update `v1` to point to the same commit
- **Sync Minor Tag**: Create or update `v1.2` to point to the same commit

This allows users of your action to reference:

- `@v1` - Always get the latest v1.x.x release
- `@v1.2` - Always get the latest v1.2.x release
- `@v1.2.3` - Pin to an exact version

**Note:** Prerelease versions (e.g., `v1.2.3-beta`, `v1.2.3-rc.1`) are skipped by default to prevent unstable versions from updating stable version tags.

## Inputs

All inputs are optional with sensible defaults for use within GitHub Actions:

- `token`: Optional - GitHub token for authentication. Defaults to `${{ github.token }}`.
- `repository`: Optional - Target repository in `owner/repo` format. Defaults to `${{ github.repository }}`.
- `git-ref`: Optional - Git reference (e.g., `refs/tags/v1.2.3`). Defaults to `${{ github.ref }}`.
- `commit-sha`: Optional - Commit SHA to point the tags to. Defaults to `${{ github.sha }}`.
- `sync-major`: Optional - Sync major version tag (e.g., `v1`). Defaults to `true`.
- `sync-minor`: Optional - Sync minor version tag (e.g., `v1.2`). Defaults to `true`.
- `skip-prereleases`: Optional - Skip syncing for prerelease versions (e.g., `v1.2.3-beta`). Defaults to `true`.
- `dry-run`: Optional - Perform a dry run without making changes. Defaults to `false`.
- `log-level`: Optional - Log level (`debug`, `info`, `warn`, `error`). Defaults to `info`.
- `github-enterprise-url`: Optional - Base URL for GitHub Enterprise (if applicable).

## Workflow Usage

Add this workflow to your repository to automatically sync version tags on every release. **No configuration required** - the action auto-discovers everything from the GitHub context:

```yaml
name: Sync Version Tags

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  sync-tags:
    name: Sync Version Tags
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: cbrgm/semver-tag-sync-action@v1
```

That's it! The action automatically uses `github.token`, `github.repository`, `github.ref`, and `github.sha` from the workflow context.

### Sync Only Major Version

If you only want to sync the major version tag (like the original `update-majorver` action):

```yaml
name: Sync Major Version Tag

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  sync-tags:
    name: Sync Major Version Tag
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: cbrgm/semver-tag-sync-action@v1
        with:
          sync-minor: false
```

### Sync Only Minor Version

If you only want to sync the minor version tag:

```yaml
name: Sync Minor Version Tag

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  sync-tags:
    name: Sync Minor Version Tag
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: cbrgm/semver-tag-sync-action@v1
        with:
          sync-major: false
```

### Include Prerelease Versions

By default, prerelease versions (e.g., `v1.2.3-beta`, `v1.2.3-rc.1`) are skipped. To also sync tags for prereleases:

```yaml
name: Sync Version Tags (Including Prereleases)

on:
  push:
    tags:
      - 'v*.*.*'
      - 'v*.*.*-*'

jobs:
  sync-tags:
    name: Sync Version Tags
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: cbrgm/semver-tag-sync-action@v1
        with:
          skip-prereleases: false
```

### Dry Run Mode

Test the action without making any changes:

```yaml
name: Test Tag Sync

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  test-sync:
    name: Test Tag Sync
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: cbrgm/semver-tag-sync-action@v1
        with:
          dry-run: true
```

### Cross-Repository Sync

Sync tags to a different repository (requires a PAT with `contents: write` permission on the target repo):

```yaml
name: Sync Tags to Another Repo

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  sync-tags:
    name: Sync Tags
    runs-on: ubuntu-latest
    steps:
      - uses: cbrgm/semver-tag-sync-action@v1
        with:
          token: ${{ secrets.PAT_TOKEN }}
          repository: owner/other-repo
```
## Container Usage

You can also run the action as a standalone container:

```bash
podman run --rm -it ghcr.io/cbrgm/semver-tag-sync-action:v1 --help
```

When running outside GitHub Actions, provide the required parameters:

```bash
podman run --rm -it ghcr.io/cbrgm/semver-tag-sync-action:v1 \
  --github-token="${GITHUB_TOKEN}" \
  --github-repo="owner/repo" \
  --git-ref="refs/tags/v1.2.3" \
  --commit-sha="abc123def456"
```

Or use environment variables (auto-discovered):

```bash
export GITHUB_TOKEN="your-token"
export GITHUB_REPOSITORY="owner/repo"
export GITHUB_REF="refs/tags/v1.2.3"
export GITHUB_SHA="abc123def456"

podman run --rm -it \
  -e GITHUB_TOKEN \
  -e GITHUB_REPOSITORY \
  -e GITHUB_REF \
  -e GITHUB_SHA \
  ghcr.io/cbrgm/semver-tag-sync-action:v1
```

## Local Development

Build the binary:

```bash
make build
```

Run tests:

```bash
make test
```

## Contributing & License

We welcome and value your contributions to this project! 👍 If you're interested in making improvements or adding features, please refer to our [Contributing Guide](https://github.com/cbrgm/semver-tag-sync-action/blob/main/CONTRIBUTING.md). This guide provides comprehensive instructions on how to submit changes, set up your development environment, and more.

Please note that this project is released with a [Contributor Code of Conduct](https://github.com/cbrgm/semver-tag-sync-action/blob/main/CODE_OF_CONDUCT.md). By participating in this project, you agree to abide by its terms.

This project is developed and distributed under the Apache 2.0 License. See the [LICENSE](https://github.com/cbrgm/semver-tag-sync-action/blob/main/LICENSE) file for more details.
