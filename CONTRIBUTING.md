# Contributing to git-sweep

## Development Setup

### Prerequisites

- Go 1.21+
- Git

### Build

```bash
# Build binary
make build

# Or directly
go build -o git-sweep .
```

### Install locally

```bash
make install
```

This installs to `~/.local/bin/` and creates the git alias.

### Run tests

```bash
make test
```

## Project Structure

```
git-sweep/
├── main.go                 # Entry point
├── cmd/                    # CLI commands (cobra)
│   ├── root.go            # Main branch cleanup command
│   ├── stash.go           # Stash cleanup
│   ├── tags.go            # Tags cleanup
│   ├── worktree.go        # Worktree cleanup
│   └── update.go          # Self-update
├── internal/
│   ├── config/            # Environment configuration
│   ├── git/               # Git operations
│   ├── provider/          # GitHub, GitLab, Bitbucket providers
│   ├── sweep/             # Branch analysis logic
│   ├── stash/             # Stash operations
│   ├── tags/              # Tags operations
│   ├── worktree/          # Worktree operations
│   ├── ui/                # Terminal UI (lipgloss, bubbletea)
│   └── update/            # Self-update logic
└── docs/
    └── vhs/               # VHS scripts for GIF generation
```

## Release Process

Releases are fully automated using [Conventional Commits](https://www.conventionalcommits.org/).

### Commit Messages

- `feat: add new feature` - Minor version bump (1.0.0 → 1.1.0)
- `fix: fix a bug` - Patch version bump (1.0.0 → 1.0.1)
- `feat!: breaking change` - Major version bump (1.0.0 → 2.0.0)
- `BREAKING CHANGE:` in commit body - Also triggers major bump

### How it works

1. Push commits to `main` branch
2. [release-please](https://github.com/googleapis/release-please) creates/updates a Release PR
3. When the Release PR is merged:
   - A new tag is created (e.g., `v2.1.0`)
   - [goreleaser](https://goreleaser.com/) builds binaries for all platforms
   - A GitHub Release is created with binaries attached
   - CHANGELOG.md is updated

### Test release locally

```bash
make release-local
```

This runs goreleaser in snapshot mode without publishing.

## Code Style

- Follow standard Go conventions
- Use `gofmt` / `goimports`
- Keep functions small and focused
- UI components go in `internal/ui/`

## Adding a New Subcommand

1. Create `internal/<feature>/<feature>.go` with the logic
2. Create `internal/ui/<feature>.go` with UI rendering
3. Create `cmd/<feature>.go` with the CLI command
4. Register in `cmd/root.go` with `cmd.AddCommand(New<Feature>Cmd())`

See `stash`, `tags`, or `worktree` for examples.

## Generating GIFs for Documentation

We use [VHS](https://github.com/charmbracelet/vhs) from Charmbracelet to generate terminal GIFs.

### Install VHS

```bash
# macOS
brew install charmbracelet/tap/vhs

# Go
go install github.com/charmbracelet/vhs@latest
```

### Generate GIFs

```bash
cd docs/vhs
vhs sweep.tape      # generates sweep.gif
vhs stash.tape      # generates stash.gif
vhs tags.tape       # generates tags.gif
vhs worktree.tape   # generates worktree.gif
```

### Edit tape files

Tape files are simple scripts. See [VHS documentation](https://github.com/charmbracelet/vhs) for syntax.
