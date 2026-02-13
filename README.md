# git-sweep

Safe local branch cleanup for squash-merge workflows.

## Features

- **Safe mode**: Only deletes branches with gone upstream AND merged PR/MR
- **Nuke mode**: Interactive multi-select to delete any branch, with smart suggestions
- **Multi-provider**: GitHub, GitLab (including self-hosted), and Bitbucket Cloud
- **Self-updating**: Built-in update command with optional auto-update
- **Pretty UI**: Spinners, colors, and charm-style output
- **Configurable**: Environment variables for per-project settings (great with direnv)

## Installation

### Quick install (recommended)

```bash
curl -sSL https://raw.githubusercontent.com/midnattsol/git-sweep/develop/install.sh | bash
```

This will:
- Download the latest release for your platform
- Install to `~/.local/bin/`
- Create the `git sweep` alias

### Go install

```bash
go install github.com/midnattsol/git-sweep@latest
git config --global alias.sweep '!git-sweep'
```

### From source

```bash
git clone https://github.com/midnattsol/git-sweep.git
cd git-sweep
make install
```

## Usage

### Safe mode (default)

Deletes local branches that:
1. Have an upstream that no longer exists (after `git fetch --prune`)
2. Have a merged PR/MR (requires provider token, see [Authentication](#authentication))

```bash
# Show what would be deleted, ask for confirmation, then delete
git sweep

# Only show, don't prompt or delete
git sweep --dry-run

# Delete without confirmation
git sweep --force

# Also delete candidates (upstream gone but no merged PR found)
git sweep --candidates

# Hide skipped branches in output
git sweep --brief
```

### Nuke mode

Interactive multi-select to delete any branch (except protected ones).

Branches are categorized and pre-selected based on their status:
- **Suggested**: Branches with merged PRs, or stale without upstream - pre-selected
- **Gone**: Upstream deleted but no merged PR found - pre-selected
- **Orphan**: No upstream but has recent commits
- **Active**: Has active upstream
- **Protected**: Cannot be deleted

```bash
# Interactive picker
git sweep --nuke

# Delete all non-protected without asking
git sweep --nuke --yes
```

### Self-update

```bash
# Check and update interactively
git sweep update

# Only check for updates
git sweep update --check

# Update without confirmation
git sweep update --yes

# Enable auto-update (shows notice when new version available)
git sweep update --auto-update=true

# Disable auto-update
git sweep update --auto-update=false
```

## Configuration

Configure via environment variables. Great with [direnv](https://direnv.net/) for per-project settings.

### General settings

| Variable | Default | Description |
|----------|---------|-------------|
| `GIT_SWEEP_REMOTE` | `origin` | Remote to use |
| `GIT_SWEEP_PROTECTED` | `master,main,develop,development,staging,release/**,hotfix/**` | Comma-separated protected patterns |
| `GIT_SWEEP_LIMIT` | `50` | Max PRs to scan |
| `GIT_SWEEP_STALE_DAYS` | `30` | Days without commits to consider a branch "stale" |
| `GIT_SWEEP_AUTO_UPDATE` | `false` | Enable auto-update check |
| `GIT_SWEEP_NO_COLOR` | `false` | Disable colors |

### Provider authentication

The provider is auto-detected from your git remote URL.

| Variable | Provider | Description |
|----------|----------|-------------|
| `GITHUB_TOKEN` | GitHub | Personal access token (falls back to `gh auth token`) |
| `GITLAB_TOKEN` | GitLab | Personal access token with `read_api` scope |
| `GITLAB_URL` | GitLab | Self-hosted GitLab URL (default: `https://gitlab.com`) |
| `BITBUCKET_USERNAME` | Bitbucket | Bitbucket username |
| `BITBUCKET_APP_PASSWORD` | Bitbucket | Bitbucket app password with `pullrequest:read` scope |

### Example `.envrc`

```bash
# .envrc for GitHub
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export GIT_SWEEP_PROTECTED="master,main,release/**"
export GIT_SWEEP_STALE_DAYS="60"
export GIT_SWEEP_AUTO_UPDATE="true"
```

```bash
# .envrc for GitLab (self-hosted)
export GITLAB_TOKEN="glpat-xxxxxxxxxxxx"
export GITLAB_URL="https://gitlab.mycompany.com"
export GIT_SWEEP_PROTECTED="master,main,release/**"
```

```bash
# .envrc for Bitbucket
export BITBUCKET_USERNAME="myuser"
export BITBUCKET_APP_PASSWORD="xxxxxxxxxxxx"
export GIT_SWEEP_PROTECTED="master,main"
```

## Authentication

`git-sweep` auto-detects the provider from your git remote URL and needs an appropriate token.

### GitHub

1. `GITHUB_TOKEN` environment variable (recommended)
2. Falls back to `gh auth token` if `gh` CLI is installed

To create a token: [github.com/settings/tokens](https://github.com/settings/tokens) with `repo` scope.

### GitLab

Set `GITLAB_TOKEN` with a personal access token with `read_api` scope.

For self-hosted GitLab, also set `GITLAB_URL` (e.g., `https://gitlab.mycompany.com`).

To create a token: GitLab > User Settings > Access Tokens.

### Bitbucket Cloud

Set both `BITBUCKET_USERNAME` and `BITBUCKET_APP_PASSWORD`.

To create an app password: Bitbucket > Personal Settings > App passwords with `Repositories: Read` and `Pull requests: Read` permissions.

## Protected Branch Patterns

- `master`, `main` - exact match
- `release/**` - matches `release/1.0`, `release/v2/hotfix`, etc.
- `release/*` - matches `release/1.0` but not `release/v2/hotfix`

## Development

```bash
# Build
make build

# Install locally
make install

# Run tests
make test

# Test release locally
make release-local
```

## Releases

Releases are automated using [Conventional Commits](https://www.conventionalcommits.org/):

- `feat: ...` - Minor version bump
- `fix: ...` - Patch version bump
- `feat!: ...` or `BREAKING CHANGE:` - Major version bump

When commits are pushed to `develop`, [release-please](https://github.com/googleapis/release-please) creates a release PR. Merging the PR triggers [goreleaser](https://goreleaser.com/) to build binaries and create a GitHub release.

## Requirements

- `git` CLI
- Provider token (GitHub, GitLab, or Bitbucket - see Authentication section)

## License

MIT
