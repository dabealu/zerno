require("gitsigns").setup({
  on_attach = function(bufnr)
    local gs = require("gitsigns")
    local map = function(keys, func, desc)
      vim.keymap.set("n", keys, func, { buffer = bufnr, desc = desc })
    end

    -- Hunk navigation
    map("]h", gs.next_hunk, "Next git hunk")
    map("[h", gs.prev_hunk, "Previous git hunk")

    -- Hunk actions (leader group, discoverable via which-key)
    vim.keymap.set("n", "<leader>gs", gs.stage_hunk, { buffer = bufnr, desc = "Stage hunk" })
    vim.keymap.set("n", "<leader>gu", gs.undo_stage_hunk, { buffer = bufnr, desc = "Undo stage hunk" })
    vim.keymap.set("n", "<leader>gr", gs.reset_hunk, { buffer = bufnr, desc = "Reset hunk" })
    vim.keymap.set("n", "<leader>gp", gs.preview_hunk, { buffer = bufnr, desc = "Preview hunk" })
    vim.keymap.set("n", "<leader>gb", gs.blame_line, { buffer = bufnr, desc = "Blame line" })
    vim.keymap.set("n", "<leader>gd", gs.diffthis, { buffer = bufnr, desc = "Diff this file" })
    -- Hunk navigation also in leader group (for discoverability)
    vim.keymap.set("n", "<leader>gn", gs.next_hunk, { buffer = bufnr, desc = "Next hunk" })
    vim.keymap.set("n", "<leader>gN", gs.prev_hunk, { buffer = bufnr, desc = "Previous hunk" })
  end,
})
