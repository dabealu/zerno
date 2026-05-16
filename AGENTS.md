# Zerno Agent Guidelines

## Project Overview

Zerno is a Go tool for automated Arch Linux installation with Sway window manager.

## Key Features

- **Standalone binary** - all configs/templates embedded via `//go:embed`
- **Two-phase installation** - `install-base` (Phase 1, chroot) + `install-full` (Phase 2, after reboot)
- **Re-run `install-full` to sync** - re-running Phase 2 syncs configs and packages
- **CachyOS support** - `cachyos` command to enable repos and kernel
- **No repo required during install** - binary is self-contained

## Key Packages

### internal/paths

**Always use this package for path resolution.** Don't hardcode paths.

### internal/assets

Embedded assets (configs, templates). Files in `assets/` are embedded at compile time.
Templates use Go's `text/template` and receive `*config.Config` for substitution:

### internal/steps

Contains function to perform operations with files and run commands

### internal/config

Config struct and loading. Saved to `~/.zerno/parameters.json`.

## Code Conventions

### Task Functions

Each task is a function returning a `Task` struct with cfg parameter:

```go
func myTask(cfg *config.Config) Task {
    return Task{
        Name: "task_name",
        RunFunc: func(cfg *config.Config) error {
            // implementation
            return nil
        },
    }
}
```

Tasks are executed via `runTaskList(tasks, cfg)`.

### Adding Config Files and Templates

1. Add file/template to `assets/` directory
2. Use `assets.Restore("path/in/assets", "/destination/path")` or `assets.RestoreTemplate("path/in/assets", "/destination/path", cfg)`
3. For directories (like `nvim/`), use `assets.RestoreDir("path/in/assets", "/destination/path")`

### Utils Source Files

Utilities are stored as `.embed` files in `assets/utilsfs/`. They are embedded and compiled at runtime.

## Commands

```bash
./build.sh all      # fmt, vet, test, build
./build.sh test     # run tests
./build.sh build    # build binary
./build.sh clean    # cleanup
```

## Available Commands

| Command | Alias | Description |
|---------|-------|-------------|
| install-base | b | (Phase 1) Run base system installation (chroot stage) |
| install-full | i | (Phase 2) Desktop/full installation (after reboot, re-run to sync) |
| qemu | q | Install and configure qemu/kvm |
| cachyos | c | (sudo) Enable CachyOS repos and kernel |
| update-bin | u | Compile new bin from local repo |
| build-iso | m | Create iso with zerno bin included |
| boot-dev | f | Format device creating storage and boot partitions |
| steam | e | (sudo) Install steam, vga: intel, nvidia, amd |
| version | v | Print version and exit |
| repo-pull | r | Clone or update repo in ~/src/zerno |

## Testing

- Integration tests use real filesystem in temp dirs
- Use `internal/testutils` for `TempFile()`, `TempDir()`, `WriteFile()` helpers
- Set `HOME` env var for tests needing config directories

## Neovim Config

Neovim configuration is in `assets/nvim/` (embedded, deployed via `install-full`).
See [vim.md](vim.md) for plugin docs, keybindings, and [vim-cheatsheet.md](vim-cheatsheet.md) for a quick reference.

## Design Decisions

- **No external dependencies** - use stdlib where possible
- **No Makefile** - use `build.sh` instead
- **Binary in repo root** - `zerno`
- **Assets embedded** - configs/templates embedded in binary
- **Repo optional** - only needed for `update-bin`, `build-iso`, `repo-pull`
