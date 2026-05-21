local bmap = require("config.utils").map

-- Reflection commands
bmap("n", "<leader>vm", "<cmd>messages<CR>", "Show messages")
bmap("n", "<leader>vh", "<cmd>checkhealth<CR>", "Health check")
bmap("n", "<leader>vM", "<cmd>Mason<CR>", "Mason LSP manager")
bmap("n", "<leader>vs", "<cmd>scriptnames<CR>", "Script names")
bmap("n", "<leader>vr", "<cmd>registers<CR>", "Registers")
bmap("n", "<leader>vc", "<cmd>history<CR>", "Command history")
bmap("n", "<leader>vv", "<cmd>version<CR>", "Version info")

bmap("n", "<leader>vn", function() Snacks.notifier.show_history() end, "Notification history")

-- Disable bigfile mode for current buffer
bmap("n", "<leader>vB", function()
  vim.cmd("filetype detect")
  vim.cmd("edit!")
  vim.schedule(function()
    pcall(vim.cmd, "LspStart")
  end)
  vim.notify("Bigfile mode disabled for this buffer", vim.log.levels.INFO)
end, "Disable bigfile mode")

-- Open vim cheatsheet
bmap("n", "<leader>v?", function()
  local path = vim.fn.expand("~/src/zerno/vim-cheatsheet.md")
  if vim.fn.filereadable(path) == 1 then
    vim.cmd("edit " .. path)
  else
    vim.notify("vim-cheatsheet.md not found at " .. path, vim.log.levels.WARN)
  end
end, "Open vim cheatsheet")
