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
