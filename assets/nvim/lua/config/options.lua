local opt = vim.opt

-- Line numbers
opt.number = true
opt.relativenumber = true

-- Whitespace rendering (matches VSCode "editor.renderWhitespace": "all")
opt.list = true
opt.listchars = { tab = "→ ", trail = "·", nbsp = "␣" }

-- Indentation
opt.expandtab = true
opt.shiftwidth = 2
opt.tabstop = 2
opt.smartindent = true

-- Search
opt.ignorecase = true
opt.smartcase = true
opt.hlsearch = true
opt.incsearch = true

-- UI
opt.signcolumn = "yes"
opt.cursorline = true
opt.termguicolors = true
opt.scrolloff = 8
opt.sidescrolloff = 8
opt.wrap = true
opt.showmode = false
opt.colorcolumn = "120"

-- Terminal title (shows in iTerm2/Alacritty tab header)
opt.title = true
opt.titlestring = "%{&modified ? '● ' : '○ '}%{fnamemodify(getcwd(), ':t')} - %t"

-- Hide command line (commands appear in popup instead)
opt.cmdheight = 0
opt.showcmdloc = "statusline"

-- Splits
opt.splitright = true
opt.splitbelow = true

-- System clipboard
opt.clipboard = "unnamedplus"

-- Persistent undo (survives closing and reopening files)
opt.undofile = true

-- Mouse support (allows clicking to position cursor if mouse reporting is enabled in terminal)
opt.mouse = "a"

-- Faster update time (default 4000ms is too slow for gitsigns and LSP)
opt.updatetime = 250
opt.timeoutlen = 750

-- Nerd Font
vim.g.have_nerd_font = true
