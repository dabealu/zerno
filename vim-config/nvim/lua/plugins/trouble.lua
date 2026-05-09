require("trouble").setup({})

local bmap = require("config.utils").map

bmap("n", "<leader>xx", "<cmd>Trouble diagnostics toggle<CR>", "Diagnostics (Trouble)")
bmap("n", "<leader>xX", "<cmd>Trouble diagnostics toggle filter.buf=0<CR>", "Buffer diagnostics (Trouble)")
bmap("n", "<leader>cs", "<cmd>Trouble symbols toggle focus=false<CR>", "Document symbols (Trouble)")
bmap("n", "<leader>cl", "<cmd>Trouble lsp toggle focus=false win.position=right<CR>", "LSP references/defs")
