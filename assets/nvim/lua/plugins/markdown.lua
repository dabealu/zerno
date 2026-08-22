require("render-markdown").setup({
  completions = { lsp = { enabled = true } },
})

local map = vim.keymap.set

map("n", "<leader>mt", "<cmd>RenderMarkdown toggle<CR>", { desc = "Markdown render toggle" })
map("n", "<leader>mp", "<cmd>RenderMarkdown preview<CR>", { desc = "Markdown side preview" })
