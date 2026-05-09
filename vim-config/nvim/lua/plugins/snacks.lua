require("snacks").setup({
  picker = {
    enabled = true,
  },
  explorer = {
    enabled = true,
  },
  notifier = {
    enabled = true,
    timeout = 10000,
  },
  indent = {
    enabled = true,
  },
})

local bmap = require("config.utils").map

-- Find group
bmap("n", "<leader>ff", function() Snacks.picker.files() end, "Find files")
bmap("n", "<leader>fg", function() Snacks.picker.grep() end, "Grep across project")
bmap("n", "<leader>fw", function() Snacks.picker.grep_word() end, "Grep word under cursor")
bmap("n", "<leader>fb", function() Snacks.picker.buffers() end, "Find buffer")
bmap("n", "<leader>fr", function() Snacks.picker.recent() end, "Recent files")
bmap("n", "<leader>f/", function() Snacks.picker.lines() end, "Search in current buffer")
bmap("n", "<leader>fh", function() Snacks.picker.help() end, "Help tags")
bmap("n", "<leader>fn", function() Snacks.notifier.show_history() end, "Notification history")

-- File explorer
bmap("n", "<leader>e", function() Snacks.explorer() end, "Toggle file explorer")

-- Project picker (scans ~/src/ for repos)
bmap("n", "<leader>p", function() Snacks.picker.projects() end, "Open project")

-- LSP pickers (defined here because they use Snacks.picker)
bmap("n", "<leader>ls", function() Snacks.picker.lsp_symbols() end, "Document symbols")
bmap("n", "<leader>ld", function() Snacks.picker.diagnostics() end, "Diagnostics")

