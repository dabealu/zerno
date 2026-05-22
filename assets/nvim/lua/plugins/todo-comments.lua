require("todo-comments").setup({})

vim.keymap.set("n", "<leader>fT", "<cmd>TodoTrouble<CR>", { desc = "Find TODOs" })
