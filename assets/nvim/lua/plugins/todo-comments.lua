require("todo-comments").setup({})

vim.keymap.set("n", "<leader>fT", function() Snacks.picker.pick("todo_comments") end, { desc = "Find TODOs" })
