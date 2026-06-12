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

## Motions

| Key | Action |
|-----|--------|
| `w` / `W` | Next word / WORD |
| `b` / `B` | Back word / WORD |
| `e` / `E` | End of word / WORD |
| `ge` / `gE` | End of previous word / WORD |
| `0` / `$` | Start / end of line |
| `^` / `g_` | First / last non-whitespace |
| `gg` / `G` | Start / end of file |
| `{line}G` | Go to line number |
| `H` / `M` / `L` | Top / middle / bottom of screen |
| `{` / `}` | Previous / next paragraph |
| `(` / `)` | Previous / next sentence |
| `%` | Matching bracket/quote |
| `gf` / `gF` | Go to file / go to file + line |
| `C-o` / `C-i` | Jump back / forward in history |
| `'`{mark} / `` ` ``{mark} | Jump to mark (line / line+column) |
| `''` / ``` `` ``` | Jump back to previous position |

> **word** = letters, digits, underscores. **WORD** = any non-whitespace characters (e.g. `foo.bar()` is 1 WORD but 3 words).

### Character Jumps

| Key | Action |
|-----|--------|
| `f{char}` | Jump forward **to** char |
| `F{char}` | Jump backward **to** char |
| `t{char}` | Jump forward **before** char |
| `T{char}` | Jump backward **before** char |
| `;` / `,` | Repeat forward / backward |

### Screen Scrolling

| Key | Action |
|-----|--------|
| `C-f` / `C-b` | Page down / page up |
| `C-d` / `C-u` | Scroll half page down / up |
| `C-e` / `C-y` | Scroll line down / up (cursor stays) |
| `zz` / `zt` / `zb` | Cursor to middle / top / bottom of screen |

## Operators

Combine with motions or text objects: `{operator}{count}{motion}`.

| Key | Action |
|-----|--------|
| `d` | Delete |
| `c` | Change (delete + insert mode) |
| `y` | Yank (copy) |
| `>` / `<` | Indent right / left (e.g. `>G`, or `>>` / `<<` for current line) |
| `=` | Auto-indent |
| `~` | Toggle case (also `g~`) |
| `gu` / `gU` | Lowercase / uppercase |
| `!` | Filter through external command |
| `gq` | Format / hard-wrap |

Examples: `d2w` (delete 2 words), `ci(` (change inside parens), `yap` (yank a paragraph), `>G` (indent to end of file).

## Text Objects

`{operator}` + `i`(inner) / `a`(around) + `{object}`.

| Object | Targets |
|--------|---------|
| `w` / `W` | Word / WORD |
| `p` | Paragraph |
| `s` | Sentence |
| `t` | HTML/XML tag |
| `"` `'` `` ` `` | Quoted strings |
| `(` `)` `b` | Parens / block |
| `{` `}` `B` | Braces / Block |
| `[` `]` | Brackets |
| `<` | Angle brackets |

Treesitter scope (same indentation): `ii`, `ai`, `[i`, `]i`.

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
| `J` | Join lines |
| `>>` / `<<` | Indent / dedent |
| `=` | Auto-indent (e.g. `=ap`) |
| `~` | Toggle case |
| `gcc` / `gc` + motion | Toggle comment |
| `*` / `#` | Search word forward/backward |
| `/` / `?` | Search |
| `n` / `N` | Next/prev search result |
| `C-d` / `C-u` | Scroll half page down / up |
| `zz` | Center screen |
| `:w` / `:q` / `:wq` / `:q!` | Save / quit |

## LSP

| Key | Action |
|-----|--------|
| `gd` | Go to definition |
| `grr` | Go to references |
| `gri` | Go to implementation |
| `gD` | Go to declaration |
| `grt` | Go to type definition |
| `grn` | Rename symbol |
| `gra` | Code action |
| `K` | Hover documentation |
| `C-o` / `C-i` | Jump back / forward |
| `[d` / `]d` | Previous / next diagnostic |
| `[w` / `]w` | Previous / next LSP reference highlight |

## Flash (supercharged `s`)

| Key | Action |
|-----|--------|
| `s` | Jump anywhere on screen (type chars, hit label) |
| `S` | Jump + select treesitter node |
| `r` (operator mode) | Cross-window jump (remote flash) |
| `R` (visual/operator) | Treesitter search |
| `C-s` (cmdline) | Toggle flash on `/` search |

## Surround (`nvim-surround`)

| Key | Action |
|-----|--------|
| `ys{motion}{char}` | Add surround |
| `ds{char}` | Delete surround |
| `cs{target}{replacement}` | Change surround |
| `S` (visual selection) | Surround selection with char |

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
LSP:      gd=def  grr=refs  gri=impl  grt=type  K=hover  [d/]d=diag
FLASH:    s=jump  S=treesitter  r=remote
SURROUND: ys=add  ds=delete  cs=change  S=visual
MACROS:   qa=record  @a=play
SPLITS:   C-h/j/k/l
```
