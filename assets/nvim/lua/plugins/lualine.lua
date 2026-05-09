require("lualine").setup({
  options = {
    theme = "kanagawa",
  },
  sections = {
    lualine_c = { {
      "filename",
      symbols = { modified = "●" },
    } },
    lualine_x = { "diagnostics", "lsp_client_name" },
  },
})
