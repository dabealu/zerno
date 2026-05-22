-- Config via global (setup() does not exist — module is a pure API)
vim.g.opencode_opts = {
  lsp = { enabled = false },  -- keep K as pure LSP hover
}

-- Restore native increment/decrement (opencode claims these by default)
pcall(vim.keymap.del, { "n", "x" }, "<C-a>")
pcall(vim.keymap.del, { "n", "x" }, "<C-x>")

-- OpenCode keymaps
vim.keymap.set({ "n", "t" }, "<C-.>", function() require("opencode").toggle() end, { desc = "Toggle opencode" })

vim.keymap.set({ "n", "x" }, "go", function() return require("opencode").operator("@this ") end, { desc = "Add range to opencode", expr = true })
vim.keymap.set("n", "goo", function() return require("opencode").operator("@this ") .. "_" end, { desc = "Add line to opencode", expr = true })

vim.keymap.set("n", "<leader>oa", function() require("opencode").ask() end, { desc = "Ask opencode" })
vim.keymap.set("n", "<leader>os", function() require("opencode").select() end, { desc = "OpenCode menu" })
vim.keymap.set("n", "<leader>ot", function() require("opencode").toggle() end, { desc = "Toggle opencode" })
vim.keymap.set("n", "<leader>ok", function()
  require("opencode").ask("@this: ", { submit = true })
end, { desc = "Explain symbol" })
