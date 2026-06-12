require("todo-comments").setup({})

vim.keymap.set("n", "<leader>fT", "<cmd>Trouble todo toggle<CR>", { desc = "Find TODOs" })
