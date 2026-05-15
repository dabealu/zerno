local M = {}

function M.map(mode, keys, action, desc, opts)
  opts = opts or {}
  opts.desc = desc
  vim.keymap.set(mode, keys, action, opts)
end

return M
