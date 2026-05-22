require("trouble").setup({})

local map = vim.keymap.set

map("n", "<leader>cx", "<cmd>Trouble diagnostics toggle<CR>", { desc = "Diagnostics" })
map("n", "<leader>cb", "<cmd>Trouble diagnostics toggle filter.buf=0<CR>", { desc = "Buffer diagnostics" })
map("n", "<leader>cs", "<cmd>Trouble symbols toggle focus=false<CR>", { desc = "Document symbols (Trouble)" })
map("n", "<leader>cl", "<cmd>Trouble lsp toggle focus=false win.position=right<CR>", { desc = "LSP references/defs" })
