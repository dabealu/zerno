local map = vim.keymap.set

-- Clear search highlight on Esc
map("n", "<Esc>", "<cmd>nohlsearch<CR>", { desc = "Clear search highlight" })

-- Double Esc in terminal mode → normal mode
map("t", "<Esc><Esc>", "<C-\\><C-n>", { desc = "Exit terminal mode (to normal mode)" })

-- Buffer navigation
map("n", "<leader>bb", "<cmd>e #<CR>", { desc = "Switch to last buffer" })
map("n", "<leader>bd", "<cmd>bdelete<CR>", { desc = "Delete buffer" })
map("n", "<leader>bn", "<cmd>bnext<CR>", { desc = "Next buffer" })
map("n", "<leader>bp", "<cmd>bprevious<CR>", { desc = "Previous buffer" })
map("n", "<leader>bl", "<cmd>ls<CR>", { desc = "List all buffers" })
map("n", "<leader>be", "<cmd>enew<CR>", { desc = "New empty buffer" })

-- Window navigation (Ctrl+hjkl to move between splits)
map("n", "<C-h>", "<C-w>h", { desc = "Move to left split" })
map("n", "<C-j>", "<C-w>j", { desc = "Move to below split" })
map("n", "<C-k>", "<C-w>k", { desc = "Move to above split" })
map("n", "<C-l>", "<C-w>l", { desc = "Move to right split" })

-- Resize splits with Ctrl+arrows
map("n", "<C-Up>", "<cmd>resize +2<CR>", { desc = "Increase split height" })
map("n", "<C-Down>", "<cmd>resize -2<CR>", { desc = "Decrease split height" })
map("n", "<C-Left>", "<cmd>vertical resize -2<CR>", { desc = "Decrease split width" })
map("n", "<C-Right>", "<cmd>vertical resize +2<CR>", { desc = "Increase split width" })

-- Move selected lines up/down in visual mode
map("v", "J", ":move '>+1<CR>gv=gv", { desc = "Move selection down" })
map("v", "K", ":move '<-2<CR>gv=gv", { desc = "Move selection up" })

-- Keep cursor centered when scrolling
map("n", "<C-d>", "<C-d>zz", { desc = "Scroll down (centered)" })
map("n", "<C-u>", "<C-u>zz", { desc = "Scroll up (centered)" })

-- Keep cursor centered when searching
map("n", "n", "nzzzv", { desc = "Next search result (centered)" })
map("n", "N", "Nzzzv", { desc = "Previous search result (centered)" })

-- Built-in LSP defaults, remapped only to provide concise which-key labels
map({ "n", "x" }, "gra", vim.lsp.buf.code_action, { desc = "Code action" })
map("n", "grd", vim.lsp.buf.definition, { desc = "Go to definition" })
map("n", "grl", vim.lsp.buf.declaration, { desc = "Go to declaration" })
map("n", "gri", vim.lsp.buf.implementation, { desc = "Go to implementation" })
map("n", "grn", vim.lsp.buf.rename, { desc = "Rename symbol" })
map("n", "grr", vim.lsp.buf.references, { desc = "Go to references" })
map("n", "grt", vim.lsp.buf.type_definition, { desc = "Go to type definition" })
map("n", "grx", vim.lsp.codelens.run, { desc = "Run code lens" })
map("n", "grk", vim.lsp.buf.hover, { desc = "Hover documentation" })

-- Better paste in visual mode (don't yank replaced text)
map("x", "p", [["_dP]], { desc = "Paste without yanking" })

-- Toggles
map("n", "<leader>tw", "<cmd>set wrap!<CR>", { desc = "Toggle line wrap" })
map("n", "<leader>ti", "<cmd>set list!<CR>", { desc = "Toggle invisible characters" })
map("n", "<leader>ts", "<cmd>set spell!<CR>", { desc = "Toggle spell check" })
map("n", "<leader>th", "<cmd>set hlsearch!<CR>", { desc = "Toggle search highlight" })
map("n", "<leader>tz", function() Snacks.zen() end, { desc = "Toggle zen mode" })

-- Insert mode: terminal-style editing
map("i", "<C-k>", '<C-o>"_D', { desc = "Delete to end of line" })
map("i", "<C-a>", "<Home>", { desc = "Jump to start of line" })
map("i", "<C-e>", "<End>", { desc = "Jump to end of line" })
map("i", "<C-Left>", "<C-o>b", { desc = "Jump word backward" })
map("i", "<C-Right>", "<C-o>w", { desc = "Jump word forward" })
map("i", "<M-b>", "<C-o>b", { desc = "Jump word backward (iTerm2)" })
map("i", "<M-f>", "<C-o>w", { desc = "Jump word forward (iTerm2)" })

-- Diagnostic navigation
map("n", "]d", function() vim.diagnostic.jump({ count = 1, float = true }) end, { desc = "Next diagnostic" })
map("n", "[d", function() vim.diagnostic.jump({ count = -1, float = true }) end, { desc = "Previous diagnostic" })

-- Oil.nvim: open parent directory
map("n", "-", "<cmd>Oil<CR>", { desc = "Open parent directory (Oil)" })

-- Maximize/unmaximize current window via tab split
map("n", "<C-w>m", function()
  if vim.t.maximized then
    vim.t.maximized = nil
    vim.cmd("tabclose")
  else
    vim.cmd("tab split")
    vim.t.maximized = true
  end
end, { desc = "Toggle maximize window" })
