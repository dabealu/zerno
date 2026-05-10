local map = vim.keymap.set
local bmap = require("config.utils").map

-- Clear search highlight on Esc
map("n", "<Esc>", "<cmd>nohlsearch<CR>", { desc = "Clear search highlight" })

-- Double Esc in terminal mode → normal mode
map("t", "<Esc><Esc>", "<C-\\><C-n>", { desc = "Exit terminal mode (to normal mode)" })

-- Buffer navigation
bmap("n", "<leader>bb", "<cmd>e #<CR>", "Switch to last buffer")
bmap("n", "<leader>bd", "<cmd>bdelete<CR>", "Delete buffer")
bmap("n", "<leader>bn", "<cmd>bnext<CR>", "Next buffer")
bmap("n", "<leader>bp", "<cmd>bprevious<CR>", "Previous buffer")
bmap("n", "<leader>bl", "<cmd>ls<CR>", "List all buffers")
bmap("n", "<leader>be", "<cmd>enew<CR>", "New empty buffer")

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

-- Better paste in visual mode (don't yank replaced text)
map("x", "p", [["_dP]], { desc = "Paste without yanking" })

-- Toggles
bmap("n", "<leader>tw", "<cmd>set wrap!<CR>", "Toggle line wrap")
bmap("n", "<leader>ti", "<cmd>set list!<CR>", "Toggle invisible characters")
bmap("n", "<leader>tc", function()
  local enabled = vim.wo.number
  vim.wo.number = not enabled
  vim.wo.relativenumber = not enabled
  vim.wo.signcolumn = enabled and "no" or "yes"
  vim.opt_local.list = not enabled
end, "Toggle copy mode (strip UI)")

-- Insert mode: terminal-style editing
map("i", "<C-k>", '<C-o>"_D', { desc = "Delete to end of line" })
map("i", "<C-a>", "<Home>", { desc = "Jump to start of line" })
map("i", "<C-e>", "<End>", { desc = "Jump to end of line" })
map("i", "<C-Left>", "<C-o>b", { desc = "Jump word backward" })
map("i", "<C-Right>", "<C-o>w", { desc = "Jump word forward" })
map("i", "<M-b>", "<C-o>b", { desc = "Jump word backward (iTerm2)" })
map("i", "<M-f>", "<C-o>w", { desc = "Jump word forward (iTerm2)" })

-- Diagnostic navigation
map("n", "]d", vim.diagnostic.goto_next, { desc = "Next diagnostic" })
map("n", "[d", vim.diagnostic.goto_prev, { desc = "Previous diagnostic" })

-- Oil.nvim: open parent directory
map("n", "-", "<cmd>Oil<CR>", { desc = "Open parent directory (Oil)" })

-- Maximize/unmaximize current window via tab split
map("n", "<C-w>m", function()
  if vim.g.maximized then
    vim.cmd("tabclose")
    vim.g.maximized = nil
  else
    vim.g.maximized = true
    vim.cmd("tab split")
  end
end, { desc = "Toggle maximize window" })
