-- Auto-switch keyboard layout: English in Normal/Cmdline, Russian in Insert.
-- The Sway config sets xkb_layout us,ru with grp:alt_shift_toggle, so index 0 = US, index 1 = RU.
-- On InsertLeave/CmdlineLeave: save current layout, switch to US (0).
-- On InsertEnter: restore the saved layout.
local function get_layout()
  local result = vim.fn.system({ "swaymsg", "-t", "get_inputs" })
  local ok, inputs = pcall(vim.json.decode, result)
  if not ok then return "0" end
  for _, input in ipairs(inputs) do
    if input.type == "keyboard" then
      return tostring(input.xkb_active_layout_index)
    end
  end
  return "0"
end

local function set_layout(idx)
  vim.fn.system({ "swaymsg", "input", "type:keyboard", "xkb_switch_layout", idx })
end

if vim.fn.executable("swaymsg") ~= 1 then
  vim.notify("[auto_layout] swaymsg not found, layout switching disabled", vim.log.levels.WARN)
  return
end

local group = vim.api.nvim_create_augroup("sway-im-select", { clear = true })

vim.api.nvim_create_autocmd({ "InsertLeave", "CmdlineLeave" }, {
  group = group,
  callback = function()
    local current = get_layout()
    vim.g.im_select_saved = current
    if current ~= "0" then
      set_layout("0")
    end
  end,
})

vim.api.nvim_create_autocmd("InsertEnter", {
  group = group,
  callback = function()
    local saved = vim.g.im_select_saved
    if saved and saved ~= "0" then
      set_layout(saved)
    end
  end,
})
