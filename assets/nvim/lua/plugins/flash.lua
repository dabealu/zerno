local map = vim.keymap.set

require("flash").setup({
  modes = {
    char = {
      enabled = true,
      jump_labels = true,
    },
  },
})

map({ "n", "x", "o" }, "s", function() require("flash").jump() end, { desc = "Flash jump" })
map({ "n", "x", "o" }, "S", function() require("flash").treesitter() end, { desc = "Flash treesitter" })
map("o", "r", function() require("flash").remote() end, { desc = "Flash remote" })
map({ "o", "x" }, "R", function() require("flash").treesitter_search() end, { desc = "Flash treesitter search" })
map("c", "<c-s>", function() require("flash").toggle() end, { desc = "Toggle flash search" })
