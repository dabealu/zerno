-- Set leader key BEFORE loading anything (required by which-key and keymaps)
vim.g.mapleader = " "
vim.g.maplocalleader = " "

-- Load core settings
require("config.options")
require("config.keymaps")
require("config.autocmds")

-- Install plugins via native vim.pack (Neovim 0.12+)
-- "~N" = major version pin (N.0.0 up to but not including N+1.0.0)
vim.pack.add({
  { src = "https://github.com/folke/snacks.nvim", version = vim.version.range("~2") },
  { src = "https://github.com/neovim/nvim-lspconfig", version = vim.version.range("~2") },
  { src = "https://github.com/mason-org/mason.nvim", version = vim.version.range("~2") },
  { src = "https://github.com/mason-org/mason-lspconfig.nvim", version = vim.version.range("~2") },
  { src = "https://github.com/saghen/blink.cmp", version = vim.version.range("~1") },
  { src = "https://github.com/folke/which-key.nvim", version = vim.version.range("~3") },
  { src = "https://github.com/lewis6991/gitsigns.nvim", version = vim.version.range("~2") },
  { src = "https://github.com/stevearc/oil.nvim", version = vim.version.range("~2") },
  { src = "https://github.com/kylechui/nvim-surround", version = vim.version.range("~4") },
  { src = "https://github.com/nvim-tree/nvim-web-devicons" },
  { src = "https://github.com/petertriho/nvim-scrollbar" }, -- no tags available
  { src = "https://github.com/Bekaboo/dropbar.nvim", version = vim.version.range("~14") },
  { src = "https://github.com/rebelot/kanagawa.nvim" },
  { src = "https://github.com/nvim-lualine/lualine.nvim" },
  { src = "https://github.com/folke/trouble.nvim" },
})

-- Load plugin configurations
require("plugins.snacks")
require("plugins.lsp")
require("plugins.git")
require("plugins.oil")
require("plugins.surround")
require("plugins.scrollbar")
require("plugins.dropbar")
require("plugins.lualine")
require("plugins.trouble")
require("plugins.ui")
