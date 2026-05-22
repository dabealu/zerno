require("snacks").setup({
  picker = {
    enabled = true,
    layout = { fullscreen = true },
    sources = {
      files = { hidden = true },
      grep = { hidden = true },
      explorer = {
        jump = { close = true }, -- Close explorer when opening files (directories stay open)
      },
      lsp_symbols = {
        filter = {
          default = {
            "Class",
            "Constructor",
            "Enum",
            "Field",
            "Function",
            "Interface",
            "Method",
            "Module",
            "Namespace",
            "Package",
            "Property",
            "Struct",
            "Trait",
            "Variable",
            "Constant",
          },
        },
      },
    },
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
  scope = {
    enabled = true,
  },
  words = {
    enabled = true,
  },
  bigfile = {
    enabled = true,
  },
  statuscolumn = {
    enabled = true,
  },
  gitbrowse = {
    enabled = true,
  },
})

local map = vim.keymap.set

-- Find group
map("n", "<leader>ff", function() Snacks.picker.files() end, { desc = "Find files" })
map("n", "<leader>fg", function() Snacks.picker.grep() end, { desc = "Grep across project" })
map("n", "<leader>fw", function() Snacks.picker.grep_word() end, { desc = "Grep word under cursor" })
map("n", "<leader>fb", function() Snacks.picker.buffers() end, { desc = "Find buffer" })
map("n", "<leader>fr", function() Snacks.picker.recent() end, { desc = "Recent files" })
map("n", "<leader>f/", function() Snacks.picker.lines() end, { desc = "Search in current buffer" })
map("n", "<leader>fh", function() Snacks.picker.help() end, { desc = "Help tags" })
map("n", "<leader>fG", function() Snacks.picker.git_status() end, { desc = "Git status (changed files)" })
map("n", "<leader>fL", function() Snacks.picker.git_log() end, { desc = "Git log (commits)" })

-- File explorer (fullscreen, closes on file open, stays open for directories)
map("n", "<leader>e", function() Snacks.picker.explorer() end, { desc = "Toggle file explorer" })

-- Words navigation (LSP references)
map("n", "]w", function() Snacks.words.jump(vim.v.count1) end, { desc = "Next LSP reference" })
map("n", "[w", function() Snacks.words.jump(-vim.v.count1) end, { desc = "Prev LSP reference" })

-- Git: open in browser
map("n", "<leader>gB", function() Snacks.gitbrowse() end, { desc = "Open in browser" })

-- Project picker (scans ~/src/ for repos)
map("n", "<leader>p", function() Snacks.picker.projects() end, { desc = "Open project" })

-- LSP pickers (defined here because they use Snacks.picker)
map("n", "<leader>ls", function() Snacks.picker.lsp_symbols() end, { desc = "Document symbols" })
map("n", "<leader>lS", function() Snacks.picker.lsp_workspace_symbols() end, { desc = "Workspace symbols" })
map("n", "<leader>ld", function() Snacks.picker.diagnostics() end, { desc = "Diagnostics" })
map("n", "<leader>lt", function()
  vim.ui.input({ prompt = "Set filetype: " }, function(input)
    if input and input ~= "" then vim.bo.filetype = input end
  end)
end, { desc = "Set filetype" })

