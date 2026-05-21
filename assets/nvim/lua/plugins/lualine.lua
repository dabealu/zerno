require("lualine").setup({
  options = {
    theme = "auto",
  },
  sections = {
    lualine_a = { "mode" },
    lualine_b = { "branch", "diff" },
    lualine_c = { {
      "filename",
      symbols = { modified = "●" },
    }, "%S" },
    lualine_x = { "filetype", "diagnostics", {
      "lsp_status",
      icon = "",
      symbols = {
        spinner = { "⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏" },
        done = "",
        separator = " "
      },
      show_name = true,
    } },
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

