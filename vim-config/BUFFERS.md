# Buffers, Windows, and Tabs

## The Three Concepts

### Buffer = an open file

Defaults:
```
:ls              Show all open buffers
:b <name>        Switch to buffer by partial name (Tab to autocomplete)
:bd              Close (delete) current buffer
```

Config:
- `<leader>fb` opens a fuzzy-searchable list of all buffers (your "tab switcher")
- `[b` / `]b` cycles through buffers (like Ctrl+Tab in VSCode)
- `<leader>bd` closes the current buffer

### Window = a visible area showing a buffer

A window is a viewport into a buffer.
```
<C-w>v           Split vertically (side by side)
<C-w>s           Split horizontally (stacked)
<C-w>h/j/k/l    Move between windows
<C-w>q           Close current window
```

`Ctrl+h/j/k/l` moves between windows (no need for `<C-w>` prefix).

### Tab = a layout of windows

It's a separate workspace layout — a collection of windows.
`:tabnew` creates a new tab page, `gt`/`gT` switches between them.

## Daily Workflow

### Open a file
- `<leader>ff` — fuzzy find by filename
- `<leader>e` — sidebar file explorer
- `-` — oil.nvim, navigate directories as editable buffers

### Switch to another open file
- `<leader>fb` — fuzzy search open buffers
- `[b` / `]b` — cycle through buffers in order

### See two files side by side
- Open first file, then `<C-w>v` to split, then `<leader>ff` to open second file
- Or: `<leader>e`, navigate to second file, press `<C-v>` to open in vertical split

### Browse/manage files (Oil.nvim)
- `-` — opens the parent directory of the current file as an editable buffer
- `Enter` on a file — opens it
- `-` again — goes up one more directory
- `/pattern` — search for a filename (normal vim search works here)
- Rename: edit the filename text, then `:w` to apply
- Delete: `dd` on a file line, then `:w` to apply
- Create: `o` to add a new line, type the filename, then `:w`
- `q` — close oil

### Close a file
- `<leader>bd` — close the buffer

### See openened files
- `<leader>fb` — shows all buffers with fuzzy search
- `:ls` — shows all buffers in a list

## Scrolling (No Trackpad Swipe)

Neovim doesn't use trackpad scrolling.

| Key | What it does | Think of it as... |
|-----|-------------|-------------------|
| `Ctrl+d` | Half page down | Your scroll-down gesture |
| `Ctrl+u` | Half page up | Your scroll-up gesture |
| `Ctrl+f` | Full page down | Fast scan forward |
| `Ctrl+b` | Full page up | Fast scan backward |
| `{` / `}` | Jump between code blocks | Skip to next function/section |
| `gg` / `G` | Top / bottom of file | Instant jump to extremes |
| `zz` | Center current line | Re-orient after jumping |

