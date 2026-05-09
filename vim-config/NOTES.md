# Personal Notes

## Registers: Multiple Clipboards

Vim has 26+ named registers (a-z) plus special ones. Think of them as labeled
slots you can yank/delete into and paste from independently.

**Using named registers** — prefix any yank/delete/paste with `"X` (quote + letter):

| Command | What it does |
|---------|-------------|
| `"ayy` | Yank line into register `a` |
| `"ap` | Paste from register `a` |
| `"bdiw` | Delete word into register `b` |
| `:reg` | Show all register contents |

**Special registers worth knowing:**

| Register | Contents | Use case |
|----------|----------|----------|
| `"` | Default (last yank or delete) | Normal `p` uses this |
| `0` | Last **yank only** (ignores deletes) | `"0p` after accidental `dd` overwrites clipboard |
| `+` | System clipboard | Linked via `unnamedplus` — same as Cmd+C/V |
| `_` | Black hole (discards) | `"_d` = true delete without touching clipboard |
| `/` | Last search pattern | |
| `.` | Last inserted text | |

**The "0 trick**: When you yank something, then delete other text (which overwrites
the default register), `"0p` still has your original yank. This is an alternative
to the visual-paste-without-yanking approach.

**In insert mode**: `Ctrl+r` then a register letter pastes from that register
without leaving insert mode. E.g. `Ctrl+r 0` pastes last yank, `Ctrl+r +`
pastes system clipboard.

**When to use named registers**: Rarely needed in practice. The main scenario is
collecting multiple things to paste later (e.g. `"a` gets a function name, `"b`
gets a URL, then paste each where needed). Most daily work uses just the default
register + the `"0` trick.

## Surround: Add/Change/Delete Surrounding Chars

Three operations, all from normal mode:

| Command | Before | After |
|---------|--------|-------|
| `ys iw "` | `hello` | `"hello"` |
| `ys iw }` | `hello` | `{hello}` |
| `cs " '` | `"hello"` | `'hello'` |
| `cs ( [` | `(hello)` | `[hello]` |
| `ds "` | `"hello"` | `hello` |
| `ds (` | `(hello)` | `hello` |

Pattern: **ys** (add) + text object + delimiter, **cs** (change) + old + new, **ds** (delete) + delimiter.

In visual mode: select text, then `S` + delimiter wraps selection.

Useful for Go/Terraform: `ysiw"` to quote a word, `cs"` + `` ` `` to switch quote style,
`ds{` to unwrap a block.

