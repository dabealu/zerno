require("scrollbar").setup({
  handle = {
    highlight = "ScrollbarHandle",
  },
  marks = {
    GitAdd = { text = "▌" },
    GitChange = { text = "▌" },
    GitDelete = { text = "▌" },
  },
  handlers = {
    gitsigns = true,
    search = false,
  },
})
