require("lualine").setup({
  options = {
    theme = "kanagawa",
  },
  sections = {
    lualine_a = { "mode" },
    lualine_b = { "branch", "diff" },
    lualine_c = { {
      "filename",
      symbols = { modified = "●" },
    } },
    lualine_x = { "diagnostics", "lsp_client_name" },
    lualine_y = { "progress" },
    lualine_z = { "location" },
  },
  inactive_sections = {
    lualine_a = {},
    lualine_b = {},
    lualine_c = { { "filename" } },
    lualine_x = {},
    lualine_y = {},
    lualine_z = { "location" },
  },
})
