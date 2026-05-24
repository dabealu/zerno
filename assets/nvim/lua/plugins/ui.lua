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
    { "<leader>d", group = "dropbar" },
    { "<leader>p", group = "project" },
    { "<leader>v", group = "vim" },
    { "<leader>o", group = "opencode" },
  },
})

vim.cmd.colorscheme("kanagawa-wave")

