require("trouble").setup({})

local bmap = require("config.utils").map

bmap("n", "<leader>cx", "<cmd>Trouble diagnostics toggle<CR>", "Diagnostics")
bmap("n", "<leader>cb", "<cmd>Trouble diagnostics toggle filter.buf=0<CR>", "Buffer diagnostics")
bmap("n", "<leader>cs", "<cmd>Trouble symbols toggle focus=false<CR>", "Document symbols (Trouble)")
bmap("n", "<leader>cl", "<cmd>Trouble lsp toggle focus=false win.position=right<CR>", "LSP references/defs")
