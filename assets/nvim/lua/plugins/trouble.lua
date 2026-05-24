require("trouble").setup({})

local map = vim.keymap.set

map("n", "<leader>lD", "<cmd>Trouble diagnostics toggle<CR>", { desc = "Diagnostics (trouble)" })
map("n", "<leader>lb", "<cmd>Trouble diagnostics toggle filter.buf=0<CR>", { desc = "Buffer diagnostics (trouble)" })
map("n", "<leader>lS", "<cmd>Trouble symbols toggle focus=false<CR>", { desc = "Document symbols (trouble)" })
map("n", "<leader>ll", "<cmd>Trouble lsp toggle focus=false win.position=right<CR>", { desc = "LSP references/defs (trouble)" })
