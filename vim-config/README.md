# Neovim Config

Neovim configuration targeting Go, Terraform, Bash, YAML.

## Architecture

```
nvim/
  init.lua                  ← entry point
  lua/config/
    options.lua             ← editor settings (numbers, whitespace, clipboard, etc.)
    keymaps.lua             ← all keybindings (non-plugin)
    autocmds.lua            ← auto-commands (format on save, session, etc.)
    utils.lua               ← bilingual keymap helper (English + Russian)
  lua/plugins/
    snacks.lua              ← fuzzy finder, file explorer, project switcher
    lsp.lua                 ← LSP setup, Mason, autocompletion
    git.lua                 ← git gutter signs, blame, hunk operations
    oil.lua                 ← buffer-based file explorer
    surround.lua            ← add/change/delete surrounding chars
    scrollbar.lua           ← scrollbar with git/diagnostic markers
    dropbar.lua             ← breadcrumb navigation bar
    statusline.lua          ← custom statusline (branch, diagnostics, LSP)
    ui.lua                  ← which-key, colorscheme
```

### Plugin Manager: vim.pack (native)

This config uses Neovim 0.12's built-in `vim.pack` module — no external plugin manager.
Plugins are declared in `init.lua` with `vim.pack.add()`. Lockfile (`nvim-pack-lock.json`) pins exact versions.

Update plugins: `:lua vim.pack.update()`

### Plugins

| Plugin | What | Why |
|--------|------|-----|
| [snacks.nvim](https://github.com/folke/snacks.nvim) | Picker, explorer, projects, notifications | Replaces telescope + nvim-tree + project.nvim in one package |
| [nvim-lspconfig](https://github.com/neovim/nvim-lspconfig) | LSP server configs | Provides config files consumed by native `vim.lsp.config()` |
| [mason.nvim](https://github.com/mason-org/mason.nvim) | LSP server installer | Auto-installs gopls, terraform-ls, etc. |
| [mason-lspconfig](https://github.com/mason-org/mason-lspconfig.nvim) | Mason <-> lspconfig bridge | Auto-enables servers installed by Mason |
| [blink.cmp](https://github.com/saghen/blink.cmp) | Autocompletion | Zero-config, async, fast fuzzy matching |
| [which-key.nvim](https://github.com/folke/which-key.nvim) | Keybinding popup | Press Space and wait — shows all available keys |
| [gitsigns.nvim](https://github.com/lewis6991/gitsigns.nvim) | Git gutter + blame | Shows added/modified/deleted lines, inline blame |
| [oil.nvim](https://github.com/stevearc/oil.nvim) | File manager as buffer | Edit filesystem like text: rename, delete, create files |
| [nvim-surround](https://github.com/kylechui/nvim-surround) | Surround operations | Add/change/delete quotes, brackets, tags (`ys`, `cs`, `ds`) |
| [nvim-scrollbar](https://github.com/petertriho/nvim-scrollbar) | Scrollbar with markers | Shows git changes and diagnostics in scrollbar |
| [dropbar.nvim](https://github.com/Bekaboo/dropbar.nvim) | Breadcrumb bar | File path + code structure navigation at top of window |
| [github-nvim-theme](https://github.com/projekt0n/github-nvim-theme) | Colorscheme | Clean light/dark GitHub-style theme |

### LSP Servers (auto-installed by Mason)

| Server | Language |
|--------|----------|
| gopls | Go |
| terraform-ls | Terraform / HCL |
| bash-language-server | Bash |
| yaml-language-server | YAML |
| json-language-server | JSON |
| lua-language-server | Lua |

## Keybindings

Leader key is **Space**.
See [CHEATSHEET.md](CHEATSHEET.md) for the full keybinding reference.

### Most Important

| Key | Action |
|-----|--------|
| `<Space>ff` | Find file (fuzzy) |
| `<Space>fg` | Grep across project |
| `<Space>fb` | Switch between open files (buffers) |
| `<Space>e` | Toggle file explorer sidebar |
| `-` | Open directory browser (oil.nvim) |
| `<Space>p` | Switch project (scans ~/src/) |
| `gd` | Go to definition |
| `gr` | Find references |
| `gi` | Find implementations |
| `K` | Hover documentation |
| `<Space>lr` | Rename symbol |
| `<Space>la` | Code action |

## Plugin Administration

### How vim.pack works

Plugins are declared in `nvim/init.lua` in the `vim.pack.add({...})` call. Each entry
is a table with `src` (GitHub URL) and `version` (semver range). On startup, vim.pack
loads plugins from `~/.local/share/nvim/site/pack/core/opt/`. If a plugin is missing
from disk, it prompts to install.

### Key paths

| What | Path |
|------|------|
| Plugin declarations | `~/.config/nvim/init.lua` |
| Plugin files on disk | `~/.local/share/nvim/site/pack/core/opt/<plugin-name>/` |
| Lockfile (pinned commits) | `~/.config/nvim/nvim-pack-lock.json` |
| Mason LSP servers | `~/.local/share/nvim/mason/` |

### Update all plugins

Inside nvim:
```vim
:lua vim.pack.update()
```
This fetches new commits, shows a changelog.
Lockfile is automatically updated with new commit hashes.

### Update a single plugin

```vim
:lua vim.pack.update({"snacks.nvim"})
```
Pass the plugin **name** (not URL) as a list.

### Add a new plugin

1. Add an entry to `vim.pack.add({...})` in `init.lua`:
   ```lua
   { src = "https://github.com/author/plugin.nvim", version = vim.version.range("~1") },
   ```
2. Create a config file in `lua/plugins/` (or add to an existing one)
3. Add `require("plugins.newplugin")` at the bottom of `init.lua`
4. Restart nvim — it will prompt to install the new plugin

### Remove a plugin

1. Remove the entry from `vim.pack.add({...})` in `init.lua`
2. Remove the corresponding `require()` line
3. Delete or clean up the config file in `lua/plugins/`
4. Delete plugin from disk:
   ```bash
   rm -rf ~/.local/share/nvim/site/pack/core/opt/<plugin-name>
   ```
   Or from inside nvim:
   ```vim
   :lua vim.pack.del({"plugin-name"})
   ```
5. The lockfile entry is cleaned up automatically on next startup

### Nuclear reset (reinstall everything)

If plugins get into a broken state:
```bash
# Remove all plugins from disk
rm -rf ~/.local/share/nvim/site/pack/core/opt/*

# Remove lockfile (forces fresh version resolution)
rm -f ~/.config/nvim/nvim-pack-lock.json

# Restart nvim — will prompt to reinstall everything
nvim
```

### LSP server management (Mason)

Mason manages LSP servers separately from vim.pack.

```vim
:Mason              " Open Mason UI
```

Inside the Mason UI:
- `i` — install a server (cursor on the server name)
- `u` — update a server
- `U` — update all servers
- `X` — uninstall a server

LSP servers live in `~/.local/share/nvim/mason/`. To nuke and reinstall:
```bash
rm -rf ~/.local/share/nvim/mason/
```
Then restart nvim — Mason will reinstall `ensure_installed` servers automatically.

## Troubleshooting

```vim
:checkhealth           " General health check
:checkhealth vim.lsp   " LSP status
:checkhealth vim.pack  " Plugin status
:Mason                 " LSP server status
:messages              " Recent messages/errors
```

If LSP isn't working for a file:
1. Check the server is installed: `:Mason`
2. Check it's running: `:lua vim.print(vim.lsp.get_clients())`
3. Check for errors: `:LspLog`

### Clipboard

`opt.clipboard = "unnamedplus"` requires a system clipboard provider:
- **macOS**: Works out of the box (pbcopy/pbpaste)
- **Linux (Wayland/Sway)**: Install `wl-clipboard` (`pacman -S wl-clipboard`)

### Terminal keycodes

iTerm2 translates `Ctrl+Left/Right` into `<M-b>`/`<M-f>` (readline sequences).
Alacritty sends actual `<C-Left>`/`<C-Right>`.
Both are mapped in `keymaps.lua` — no switching needed.

