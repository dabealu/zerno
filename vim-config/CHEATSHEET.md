# Neovim Cheatsheet

## Phase 1

### Modes

| Key | Action |
|-----|--------|
| `i` | Enter insert mode (type text) |
| `Esc` | Return to normal mode |
| `v` | Enter visual mode (select text) |
| `V` | Enter visual line mode (select whole lines) |

### Movement (Normal Mode)

| Key | Action |
|-----|--------|
| `h` `j` `k` `l` | Left, down, up, right |
| `w` | Jump forward one word |
| `b` | Jump backward one word |
| `0` | Start of line |
| `$` | End of line |
| `gg` | Top of file |
| `G` | Bottom of file |

### Basic Editing

| Key | Action |
|-----|--------|
| `dd` | Delete entire line |
| `u` | Undo |
| `Ctrl+r` | Redo |
| `yy` | Copy (yank) entire line |
| `p` | Paste below cursor |
| `P` | Paste above cursor |
| `o` | New line below and enter insert mode |
| `O` | New line above and enter insert mode |

### Scrolling (replaces trackpad swipe)

| Key | Action |
|-----|--------|
| `Ctrl+d` | Scroll half page down (your main "scroll down") |
| `Ctrl+u` | Scroll half page up (your main "scroll up") |
| `Ctrl+f` | Full page down (faster scanning) |
| `Ctrl+b` | Full page up |
| `{` / `}` | Jump to previous / next empty line (between code blocks) |
| `zz` | Center current line on screen |

### Files and Navigation

| Key | Action |
|-----|--------|
| `:w` | Save |
| `:q` | Quit |
| `:wq` | Save and quit |
| `:q!` | Quit without saving |
| `<leader>ff` | Find file (fuzzy) |
| `<leader>e` | Toggle file explorer sidebar |
| `<leader>fg` | Search text across project |
| `-` | Open directory browser (Oil) |

### Search

| Key | Action |
|-----|--------|
| `/pattern` | Search forward |
| `?pattern` | Search backward |
| `n` | Next match |
| `N` | Previous match |
| `Esc` | Clear search highlight |

---

## Phase 2

### More Movement

| Key | Action |
|-----|--------|
| `e` | Jump to end of word |
| `{` / `}` | Jump by paragraph |
| `%` | Jump to matching bracket |
| `Ctrl+d` | Scroll half page down |
| `Ctrl+u` | Scroll half page up |
| `H` / `M` / `L` | Move cursor to top/middle/bottom of screen |

### Editing Efficiently

| Key | Action |
|-----|--------|
| `ciw` | Change inner word (delete word, enter insert mode) |
| `ci"` | Change inside quotes |
| `ci(` | Change inside parentheses |
| `ci{` | Change inside braces |
| `diw` | Delete inner word |
| `di"` | Delete inside quotes |
| `A` | Append at end of line |
| `I` | Insert at start of line |
| `cc` | Change entire line |
| `C` | Change from cursor to end of line |
| `D` | Delete from cursor to end of line |
| `x` | Delete character under cursor |
| `>>` / `<<` | Indent / dedent line |
| `gcc` | Toggle comment on current line |
| `gc` + motion | Toggle comment on block (e.g. `gcj`, `gcap`) |

> `g` is a Vim prefix key — it modifies the next key (e.g. `gc`=comment, `gq`=format, `gf`=open file). Most common `g` binds are listed throughout this cheatsheet.

### Visual Mode

| Key | Action |
|-----|--------|
| `v` + movement | Select text |
| `V` + movement | Select lines |
| `Ctrl+v` | Block (column) selection |
| `y` | Yank (copy) selection |
| `d` | Delete selection |
| `>` / `<` | Indent / dedent selection |
| `J` / `K` | Move selected lines down / up (our config) |

### Buffers

| Key | Action |
|-----|--------|
| `<leader>fb` | Fuzzy find open buffers |
| `[b` / `]b` | Previous / next buffer |
| `<leader>bd` | Close buffer |
| `:ls` | List all buffers |

### Search and Replace

| Key | Action |
|-----|--------|
| `:%s/old/new/g` | Replace all in file |
| `:%s/old/new/gc` | Replace all with confirmation |
| `:5,20s/old/new/g` | Replace in lines 5-20 |
| `*` | Search for word under cursor |
| `#` | Search backward for word under cursor |

---

## Phase 3

### Navigation

| Key | Action |
|-----|--------|
| `gd` | Go to definition |
| `gr` | Go to references |
| `gi` | Go to implementation |
| `gD` | Go to declaration |
| `gy` | Go to type definition |
| `K` | Hover documentation |
| `Ctrl+o` | Jump back (after gd/gr/gi) |
| `Ctrl+i` | Jump forward |

### Actions

| Key | Action |
|-----|--------|
| `<leader>lr` | Rename symbol (across codebase) |
| `<leader>la` | Code action (quick fixes, refactors) |
| `<leader>lf` | Format buffer |
| `<leader>ls` | Document symbols (functions, types in current file) |
| `<leader>ld` | Diagnostics list (errors, warnings) |
| `[d` / `]d` | Previous / next diagnostic |

### Git

| Key | Action |
|-----|--------|
| `[h` / `]h` | Previous / next changed hunk |
| `<leader>gs` | Stage hunk |
| `<leader>gu` | Undo stage hunk |
| `<leader>gr` | Reset hunk |
| `<leader>gp` | Preview hunk (see the diff) |
| `<leader>gb` | Blame current line |
| `<leader>gd` | Diff current file |

---

## Phase 4

### Text Objects (the `i`/`a` system)

`d`/`c`/`y`/`v` + `i`/`a` + object — the most powerful vim concept.
`i` = inner (content only), `a` = around (content + delimiters).

| Key | Action |
|-----|--------|
| `dap` | Delete a paragraph |
| `yiw` | Yank inner word |
| `ci'` | Change inside single quotes |
| `va{` | Visual select around braces |
| `di[` | Delete inside square brackets |
| `dat` | Delete around HTML tag |

### Character Jumps

| Key | Action |
|-----|--------|
| `f<char>` | Jump forward to character |
| `F<char>` | Jump backward to character |
| `t<char>` | Jump forward to before character |
| `;` | Repeat last f/F/t/T jump |
| `,` | Repeat last f/F/t/T jump (reverse) |

### The Power Keys

| Key | Action |
|-----|--------|
| `.` | Repeat last change (extremely powerful) |
| `*` | Search word under cursor |
| `~` | Toggle case of character |
| `J` | Join current line with next |
| `=` | Auto-indent (e.g., `=ap` to indent a paragraph) |

### Macros

| Key | Action |
|-----|--------|
| `qa` | Start recording macro into register `a` |
| `q` | Stop recording |
| `@a` | Play macro from register `a` |
| `@@` | Replay last macro |
| `10@a` | Play macro 10 times |

### Splits

| Key | Action |
|-----|--------|
| `<C-w>v` | Vertical split |
| `<C-w>s` | Horizontal split |
| `Ctrl+h/j/k/l` | Navigate between splits (our config) |
| `<C-w>q` | Close split |
| `<C-w>=` | Equalize split sizes |
| `Ctrl+arrows` | Resize splits (our config) |

### Oil.nvim (File Manager)

| Key | Action |
|-----|--------|
| `-` | Open parent directory |
| `Enter` | Open file/directory |
| `-` (inside oil) | Go up one directory |
| Edit filename | Rename file (then `:w` to apply) |
| `dd` on a line | Delete file (then `:w` to apply) |
| `o` + type name | Create file (then `:w` to apply) |
| `<C-v>` | Open in vertical split |
| `<C-s>` | Open in horizontal split |
| `g?` | Show help |
| `q` | Close oil |

---

## Quick Reference Card

```
MODES:  i=insert  Esc=normal  v=visual  V=visual-line  :=command

SAVE:   :w     QUIT: :q    BOTH: :wq    FORCE QUIT: :q!

MOVE:   hjkl=arrows  w/b=word  0/$=line  gg/G=file  Ctrl-d/u=page

EDIT:   dd=delete line  yy=copy line  p=paste  u=undo  Ctrl-r=redo
        ciw=change word  ci"=change in quotes  .=repeat last

FIND:   /=search  n/N=next/prev  *=search word  <leader>ff=find file
        <leader>fg=grep project  <leader>fb=find buffer

LSP:    gd=definition  gr=references  gi=implementation  K=hover
        <leader>lr=rename  <leader>la=code action  <leader>lf=format

GIT:    [h/]h=prev/next hunk  <leader>gb=blame  <leader>gp=preview

FILES:  <leader>e=explorer  -=oil  <leader>p=projects
```

