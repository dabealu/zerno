-- which-key: keybinding discoverability
require("which-key").setup({
  icons = {
    mappings = vim.g.have_nerd_font,
  },
  spec = {
    { "<leader>f", group = "find" },
    { "<leader>l", group = "lsp" },
    { "<leader>g", group = "git" },
    { "<leader>b", group = "buffer" },
    { "<leader>t", group = "toggle" },
    { "<leader>x", group = "troubleshoot" },
    { "<leader>c", group = "code" },
  },
})

-- Colorscheme: kanagawa (dragon — warm, rusty, very dark)
vim.cmd.colorscheme("kanagawa-dragon")

-- Make whitespace chars (listchars) more visible
local ws_fg = vim.o.background == "dark" and "#4a4a4a" or "#b0b0b0"
vim.api.nvim_set_hl(0, "Whitespace", { fg = ws_fg })
