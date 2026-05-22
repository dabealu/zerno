require("dropbar").setup({
  icons = {
    ui = {
      bar = {
        separator = " > ",
      },
    },
  },
})

local map = vim.keymap.set

map("n", "<leader>dp", function() require("dropbar.api").pick() end, { desc = "Pick breadcrumb item" })
map("n", "<leader>dn", function() require("dropbar.api").select_next_context() end, { desc = "Next context" })
map("n", "<leader>dN", function() require("dropbar.api").select_prev_context() end, { desc = "Prev context" })
