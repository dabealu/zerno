-- Lazy parser installation: install on first filetype open

-- markdown_inline has no filetype mapping, but render-markdown plugin needs it alongside markdown.
if not pcall(vim.treesitter.language.inspect, "markdown_inline") then
  require("nvim-treesitter").install({ "markdown_inline" })
end

vim.api.nvim_create_autocmd("FileType", {
  group = vim.api.nvim_create_augroup("treesitter-install", { clear = true }),
  callback = function(args)
    local lang = vim.treesitter.language.get_lang(args.match)
    if lang and not pcall(vim.treesitter.language.inspect, lang) then
      local parsers = require("nvim-treesitter.parsers")
      if parsers[lang] then
        require("nvim-treesitter").install({ lang })
      end
    end
  end,
})
