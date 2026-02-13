# git-sweep

Safe local branch cleanup for squash-merge workflows.

## Features

- **Safe mode**: Only deletes branches with gone upstream AND merged PR
- **Nuke mode**: Interactive multi-select to delete any branch, with smart suggestions
- **Self-updating**: Built-in update command with optional auto-update
- **Pretty UI**: Spinners, colors, and charm-style output
- **Configurable**: Environment variables for per-project settings (great with direnv)
- **No `gh` required**: Uses GitHub API directly (just needs a token)

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
2. Have a merged PR on GitHub

```bash
# Dry run (default)
git sweep

# Actually delete
git sweep --execute

# Show skip reasons with branch age
git sweep --verbose
```

### Nuke mode

Interactive multi-select to delete any branch (except protected ones).

Branches are categorized and pre-selected based on their status:
- **Suggested**: Stale branches (no upstream, no recent commits) - pre-selected
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

| Variable | Default | Description |
|----------|---------|-------------|
| `GITHUB_TOKEN` | - | GitHub token (falls back to `gh auth token`) |
| `GIT_SWEEP_REMOTE` | `origin` | Remote to use |
| `GIT_SWEEP_PROTECTED` | `master,main,develop,development,staging,release/**,hotfix/**` | Comma-separated protected patterns |
| `GIT_SWEEP_LIMIT` | `50` | Max PRs to scan |
| `GIT_SWEEP_STALE_DAYS` | `30` | Days without commits to consider a branch "stale" |
| `GIT_SWEEP_AUTO_UPDATE` | `false` | Enable auto-update check |
| `GIT_SWEEP_NO_COLOR` | `false` | Disable colors |

### Example `.envrc`

```bash
# .envrc
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export GIT_SWEEP_PROTECTED="master,main,release/**"
export GIT_SWEEP_REMOTE="upstream"
export GIT_SWEEP_STALE_DAYS="60"
export GIT_SWEEP_AUTO_UPDATE="true"
```

## Authentication

`git-sweep` needs a GitHub token to check PR status. It looks for:

1. `GITHUB_TOKEN` environment variable (recommended)
2. Falls back to `gh auth token` if `gh` CLI is installed and authenticated

To create a token:
1. Go to https://github.com/settings/tokens
2. Generate a new token with `repo` scope
3. Set it as `GITHUB_TOKEN` in your environment

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
- GitHub token (via `GITHUB_TOKEN` env var or `gh auth token`)

## License

MIT
