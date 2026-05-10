-- Clean up mangled oil:// buffers left by session restore (e.g. buffer named "://")
local augroup = vim.api.nvim_create_augroup("oil-mangled-buffer", { clear = true })
vim.api.nvim_create_autocmd("BufReadCmd", {
  group = augroup,
  pattern = "://*",
  callback = function()
    vim.schedule(function()
      pcall(vim.api.nvim_buf_delete, vim.fn.bufnr(), { force = true })
    end)
  end,
})

require("oil").setup({
  default_file_explorer = true,
  columns = {
    "icon",
    "size",
  },
  view_options = {
    show_hidden = true,
  },
  keymaps = {
    ["g?"] = "actions.show_help",
    ["<CR>"] = "actions.select",
    ["-"] = "actions.parent",
    ["<C-v>"] = "actions.select_vsplit",
    ["<C-s>"] = "actions.select_split",
    ["<C-r>"] = "actions.refresh",
    ["q"] = "actions.close",
  },
})
