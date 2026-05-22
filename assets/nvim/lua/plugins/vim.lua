local map = vim.keymap.set

-- Reflection commands
map("n", "<leader>vm", "<cmd>messages<CR>", { desc = "Show messages" })
map("n", "<leader>vh", "<cmd>checkhealth<CR>", { desc = "Health check" })
map("n", "<leader>vM", "<cmd>Mason<CR>", { desc = "Mason LSP manager" })
map("n", "<leader>vs", "<cmd>scriptnames<CR>", { desc = "Script names" })
map("n", "<leader>vr", "<cmd>registers<CR>", { desc = "Registers" })
map("n", "<leader>vc", "<cmd>history<CR>", { desc = "Command history" })
map("n", "<leader>vv", "<cmd>version<CR>", { desc = "Version info" })

map("n", "<leader>vn", function() Snacks.notifier.show_history() end, { desc = "Notification history" })

-- Disable bigfile mode for current buffer
map("n", "<leader>vB", function()
  vim.cmd("filetype detect")
  vim.cmd("edit!")
  vim.schedule(function()
    pcall(vim.cmd, "LspStart")
  end)
  vim.notify("Bigfile mode disabled for this buffer", vim.log.levels.INFO)
end, { desc = "Disable bigfile mode" })

-- Open vim cheatsheet
map("n", "<leader>v?", function()
  local path = vim.fn.expand("~/src/zerno/vim-cheatsheet.md")
  if vim.fn.filereadable(path) == 1 then
    vim.cmd("edit " .. path)
  else
    vim.notify("vim-cheatsheet.md not found at " .. path, vim.log.levels.WARN)
  end
end, { desc = "Open vim cheatsheet" })
