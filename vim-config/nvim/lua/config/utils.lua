local M = {}

-- Positional QWERTY → Russian layout mapping
local en_to_ru = {
  a = "ф", b = "и", c = "с", d = "в", e = "у", f = "а", g = "п",
  h = "р", i = "ш", j = "о", k = "л", l = "д", m = "ь", n = "т",
  o = "щ", p = "з", q = "й", r = "к", s = "ы", t = "е", u = "г",
  v = "м", w = "ц", x = "ч", y = "н", z = "я",
  A = "Ф", B = "И", C = "С", D = "В", E = "У", F = "А", G = "П",
  H = "Р", I = "Ш", J = "О", K = "Л", L = "Д", M = "Ь", N = "Т",
  O = "Щ", P = "З", Q = "Й", R = "К", S = "Ы", T = "Е", U = "Г",
  V = "М", W = "Ц", X = "Ч", Y = "Н", Z = "Я",
}

function M.map(mode, keys, action, desc, opts)
  opts = opts or {}
  opts.desc = desc
  vim.keymap.set(mode, keys, action, opts)

  -- If it's a leader mapping, also register Russian version (no desc = hidden from which-key)
  local prefix, combo = keys:match("^(<leader>)(.+)$")
  if prefix and combo then
    local ru_combo = combo:gsub(".", function(c) return en_to_ru[c] or c end)
    if ru_combo ~= combo then
      local ru_opts = vim.tbl_extend("force", opts, {})
      ru_opts.desc = nil
      vim.keymap.set(mode, prefix .. ru_combo, action, ru_opts)
    end
  end
end

return M
