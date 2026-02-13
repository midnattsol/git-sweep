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

Safe branch cleanup for squash-merge workflows.

```bash
git sweep               # Delete branches with gone upstream
git sweep --dry-run     # Only show, don't delete
git sweep --force       # Delete without confirmation
git sweep --only-merged # Only delete if PR was merged (stricter)
git sweep --brief       # Hide skipped branches
```

**What gets deleted:**
- Branches where upstream no longer exists (after `git fetch --prune`)
- Use `--only-merged` to also require a confirmed merged PR

Branches that were never pushed are ignored - your local experiments stay safe.

#### Nuke mode

Interactive picker to delete any branch (except protected). Use when you want full control.

```bash
git sweep --nuke       # Interactive multi-select
git sweep --nuke --yes # Delete all suggested without asking
```

**What gets suggested for deletion:**
- Branches with merged PRs
- Branches with gone upstream
- **Stale branches**: no upstream + last commit older than 30 days

Stale detection helps clean up old local branches you may have forgotten about. Configure the threshold with `GIT_SWEEP_STALE_DAYS`.

#### Branch categories

| Category | Description | Normal mode | Nuke mode |
|----------|-------------|-------------|-----------|
| **Gone** | Upstream deleted | ✅ Deletes | ✅ Suggested |
| **Gone + Merged** | Upstream deleted + merged PR | ✅ Deletes | ✅ Suggested |
| **Stale** | No upstream, old commits (>30d) | Ignored | ✅ Suggested |
| **Orphan** | No upstream, recent commits | Ignored | Selectable |
| **Active** | Has active upstream | Ignored | Selectable |
| **Protected** | Current or protected branch | Skipped | Skipped |

With `--only-merged`, normal mode only deletes branches that have both gone upstream AND a confirmed merged PR.

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
| `GIT_SWEEP_STALE_DAYS` | `30` | Days until a local branch is considered "stale" in nuke mode |
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
export BITBUCKET_USERNAME="your-username"
export BITBUCKET_TOKEN="xxxxxxxxxxxx"
```

Create token: Bitbucket > Personal Settings > API tokens with `Repositories: Read` and `Pull requests: Read` scopes.

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
