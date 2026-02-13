# git-sweep

All-in-one git repository cleanup: branches, stashes, tags, and worktrees.

[![Release](https://img.shields.io/github/v/release/midnattsol/git-sweep)](https://github.com/midnattsol/git-sweep/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/midnattsol/git-sweep)](go.mod)
[![License](https://img.shields.io/github/license/midnattsol/git-sweep)](LICENSE)

![git-sweep demo](docs/images/demo.gif)

## Features

- **Branches** - Safe cleanup with PR merge detection (squash-merge aware)
- **Stashes** - List and clean old stashes with context (date, branch, files)
- **Tags** - Find and delete orphan tags (local without remote)
- **Worktrees** - Detect and remove broken worktrees
- **Multi-provider** - GitHub, GitLab (self-hosted), Bitbucket Cloud
- **Beautiful UI** - Spinners, colors, interactive pickers

## Install

### Quick install (recommended)

```bash
curl -sSL https://raw.githubusercontent.com/midnattsol/git-sweep/main/install.sh | bash
```

### Go install

```bash
go install github.com/midnattsol/git-sweep@latest
```

Then create the git alias:

```bash
git config --global alias.sweep '!git-sweep'
```

## Commands

### `git sweep` - Branch cleanup

Deletes local branches that have:
1. An upstream that no longer exists (after `git fetch --prune`)
2. A merged PR/MR on the remote

```bash
git sweep              # Show eligible, confirm, delete
git sweep --dry-run    # Only show, don't delete
git sweep --force      # Delete without confirmation
git sweep --candidates # Include candidates (gone upstream, no merged PR)
git sweep --brief      # Hide skipped branches
```

#### Nuke mode

Interactive picker to delete any branch (except protected).

```bash
git sweep --nuke       # Interactive multi-select
git sweep --nuke --yes # Delete all suggested without asking
```

### `git sweep stash` - Stash cleanup

List and clean old stashes with useful context.

```bash
git sweep stash              # List stashes with date, branch, files
git sweep stash --clean      # Interactive picker to delete
git sweep stash --force      # Delete stashes older than threshold
git sweep stash --days 60    # Set age threshold (default: 30)
```

### `git sweep tags` - Tag cleanup

Find and delete orphan tags (local tags without remote counterpart).

```bash
git sweep tags           # List all tags, highlight orphans
git sweep tags --clean   # Interactive picker to delete orphans
git sweep tags --force   # Delete all orphans without asking
```

### `git sweep worktree` - Worktree cleanup

Detect and remove broken worktrees.

```bash
git sweep worktree           # List worktrees, show status
git sweep worktree --clean   # Interactive picker to remove broken
git sweep worktree --force   # Remove all broken without asking
```

### `git sweep update` - Self-update

```bash
git sweep update         # Check and update interactively
git sweep update --check # Only check for updates
git sweep update --yes   # Update without confirmation
```

### `git sweep protect` / `unprotect` - Manage protected branches

```bash
git sweep protect staging        # Add to protected list
git sweep protect "release/**"   # Patterns supported
git sweep unprotect staging      # Remove from protected list
git sweep protect --list         # Show all protected patterns
```

## Configuration

Configure via environment variables. Works great with [direnv](https://direnv.net/).

| Variable | Default | Description |
|----------|---------|-------------|
| `GIT_SWEEP_REMOTE` | `origin` | Remote to use |
| `GIT_SWEEP_PROTECTED` | `master,main,develop` | Protected branch patterns (overrides config) |
| `GIT_SWEEP_LIMIT` | `50` | Max PRs to scan |
| `GIT_SWEEP_STALE_DAYS` | `30` | Days threshold for stale detection |
| `GIT_SWEEP_NO_COLOR` | `false` | Disable colors |

### Protected branches

By default, `master`, `main`, and `develop` are protected. Add more with:

```bash
git sweep protect staging
git sweep protect "release/**"
```

Patterns support wildcards:
- `release/*` matches `release/1.0` but not `release/v2/hotfix`
- `release/**` matches `release/1.0` and `release/v2/hotfix`

Config is stored in `~/.config/git-sweep/config.yaml`. Use `GIT_SWEEP_PROTECTED` env var to override completely.

## Authentication

Provider is auto-detected from your git remote URL.

### GitHub

```bash
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
```

Create token: [github.com/settings/tokens](https://github.com/settings/tokens) with `repo` scope.

### GitLab

```bash
export GITLAB_TOKEN="glpat-xxxxxxxxxxxx"

# For self-hosted GitLab
export GITLAB_URL="https://gitlab.mycompany.com"
```

Create token: GitLab > User Settings > Access Tokens with `read_api` scope.

### Bitbucket Cloud

```bash
export BITBUCKET_TOKEN="xxxxxxxxxxxx"
```

Create token: Bitbucket > Repository Settings > Access tokens with `Read` scope.

## Example `.envrc`

```bash
# GitHub project
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export GIT_SWEEP_STALE_DAYS="60"
```

```bash
# Self-hosted GitLab
export GITLAB_TOKEN="glpat-xxxxxxxxxxxx"
export GITLAB_URL="https://gitlab.mycompany.com"
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, project structure, and release process.

## License

MIT
