require("gitsigns").setup({
  on_attach = function(bufnr)
    local gs = require("gitsigns")
    local bmap = require("config.utils").map
    local buf_opts = { buffer = bufnr }
    local map = function(mode, keys, func, desc)
      vim.keymap.set(mode, keys, func, { buffer = bufnr, desc = desc })
    end

    -- Hunk navigation (langmap handles Russian)
    map("n", "]h", gs.next_hunk, "Next git hunk")
    map("n", "[h", gs.prev_hunk, "Previous git hunk")

    -- Hunk actions (leader group — bilingual)
    bmap("n", "<leader>gs", gs.stage_hunk, "Stage hunk", buf_opts)
    bmap("n", "<leader>gu", gs.undo_stage_hunk, "Undo stage hunk", buf_opts)
    bmap("n", "<leader>gr", gs.reset_hunk, "Reset hunk", buf_opts)
    bmap("n", "<leader>gp", gs.preview_hunk, "Preview hunk", buf_opts)
    bmap("n", "<leader>gb", gs.blame_line, "Blame line", buf_opts)
    bmap("n", "<leader>gd", gs.diffthis, "Diff this file", buf_opts)
  end,
})
