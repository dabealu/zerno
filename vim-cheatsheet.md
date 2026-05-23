# Neovim Cheatsheet

## Leader Groups (`<space>` then which-key)

| Group | Key | Purpose |
|-------|-----|---------|
| Find | `f` | Files, grep, buffers, recent, git, TODOs |
| LSP | `l` | Rename, code actions, format, diagnostics, symbols, references |
| Git | `g` | Hunks, blame, diff, browser |
| Buffer | `b` | Next, prev, delete, list, switch |
| Toggle | `t` | Wrap, invis, spell, hlsearch, zen |
| Dropbar | `d` | Breadcrumb navigation |
| Project | `p` | Switch projects (scans ~/src/) |
| Vim | `v` | Messages, health, mason, registers, config |
| OpenCode | `o` | Ask, explain, menu, toggle |

## Core Vim

| Key | Action |
|-----|--------|
| `i` | Insert mode |
| `Esc` | Normal mode |
| `v` / `V` / `C-v` | Visual / line / block select |
| `u` / `C-r` | Undo / redo |
| `.` | Repeat last change |
| `dd` / `yy` / `p` / `P` | Delete, yank, paste line |
| `ciw` / `diw` / `yiw` | Change/delete/yank inner word |
| `ci"` / `ci(` / `ci{` | Change inside quotes/parens/braces |
| `>>` / `<<` | Indent / dedent |
| `J` | Join lines |
| `=` | Auto-indent (e.g. `=ap`) |
| `~` | Toggle case |
| `gcc` / `gc` + motion | Toggle comment |
| `*` / `#` | Search word forward/backward |
| `/` / `?` | Search |
| `n` / `N` | Next/prev search result |
| `:w` / `:q` / `:wq` / `:q!` | Save / quit |
| `C-d` / `C-u` | Scroll half page down/up (centered) |
| `zz` | Center screen |

## LSP

| Key | Action |
|-----|--------|
| `gd` | Go to definition |
| `gr` | Go to references |
| `gi` | Go to implementation |
| `gD` | Go to declaration |
| `gy` | Go to type definition |
| `K` | Hover documentation |
| `C-o` / `C-i` | Jump back / forward |
| `[d` / `]d` | Previous / next diagnostic |
| `[w` / `]w` | Previous / next LSP reference highlight |

## Flash (supercharged `s`)

| Key | Action |
|-----|--------|
| `s` | Jump anywhere on screen (type chars, hit label) |
| `S` | Jump to treesitter node |
| `r` (operator mode) | Cross-window jump (remote flash) |
| `R` (visual/operator) | Treesitter search |
| `C-s` (cmdline) | Toggle flash on `/` search |

## Surround (`nvim-surround`)

| Key | Action |
|-----|--------|
| `ys{motion}{char}` | Add surround |
| `ds{char}` | Delete surround |
| `cs{target}{replacement}` | Change surround |
| `gs` (visual selection) | Surround selection with char |

## Text Objects

`d`/`c`/`y`/`v` + `i`(inner)/`a`(around) + object.

| Example | Action |
|---------|--------|
| `dap` | Delete a paragraph |
| `yiw` | Yank inner word |
| `ci'` | Change inside single quotes |
| `va{` | Visual select around braces |
| `dat` | Delete around HTML tag |

Treesitter scope (same indentation): `dii`, `dai`, `vii`, `[i` / `]i`.

## Character Jumps

| Key | Action |
|-----|--------|
| `f{char}` | Jump forward to char |
| `F{char}` | Jump backward to char |
| `t{char}` | Jump forward to before char |
| `;` / `,` | Repeat forward / backward |

## Macros

| Key | Action |
|-----|--------|
| `qa` | Record into register `a` |
| `q` | Stop recording |
| `@a` | Play register `a` |
| `@@` | Repeat last macro |

## Splits

| Key | Action |
|-----|--------|
| `C-w v` / `C-w s` | Vertical / horizontal split |
| `C-h/j/k/l` | Navigate splits |
| `C-w m` | Maximize / unmaximize |
| `C-arrows` | Resize splits |
| `C-w q` | Close split |

## Oil (file manager)

| Key | Action |
|-----|--------|
| `-` | Open parent directory |
| `Enter` | Open file / directory |
| Edit filename + `:w` | Rename |
| `dd` + `:w` | Delete file |
| `o` + name + `:w` | Create file |
| `g?` | Help |

## Quick Reference

```
MODES:    i=insert  Esc=normal  v=visual  V=line  C-v=block
SAVE/QUIT:  :w  :q  :wq  :q!
LSP:      gd=def  gr=refs  gi=impl  gy=type  K=hover  [d/]d=diag
FLASH:    s=jump  S=treesitter  r=remote
SURROUND: ys=add  ds=delete  cs=change  gs=visual
MACROS:   qa=record  @a=play
SPLITS:   C-h/j/k/l

OpenCode: C-.=toggle  go=motion  goo=line
