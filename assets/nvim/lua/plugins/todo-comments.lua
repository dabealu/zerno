require("todo-comments").setup({})

local bmap = require("config.utils").map

-- Find TODOs across project (opens in trouble tree view)
bmap("n", "<leader>fT", "<cmd>TodoTrouble<CR>", "Find TODOs")
