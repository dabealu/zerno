local autocmd = vim.api.nvim_create_autocmd
local augroup = vim.api.nvim_create_augroup

-- Highlight text briefly after yanking
autocmd("TextYankPost", {
  group = augroup("highlight-yank", { clear = true }),
  callback = function()
    vim.hl.on_yank({ timeout = 200 })
  end,
})

-- Format on save (Go: organize imports first, then format)
autocmd("BufWritePre", {
  group = augroup("format-on-save", { clear = true }),
  pattern = { "*.go", "*.tf", "*.tfvars" },
  callback = function()
    if vim.bo.filetype == "go" then
      local params = vim.lsp.util.make_range_params()
      params.context = { only = { "source.organizeImports" } }
      local result = vim.lsp.buf_request_sync(0, "textDocument/codeAction", params, 1000)
      for _, res in pairs(result or {}) do
        for _, action in pairs(res.result or {}) do
          if action.edit then
            vim.lsp.util.apply_workspace_edit(action.edit, "utf-8")
          elseif action.command then
            vim.lsp.buf.execute_command(action.command)
          end
        end
      end
    end
    vim.lsp.buf.format({ async = false })
  end,
})

-- Go: use tabs (gofmt standard)
autocmd("FileType", {
  group = augroup("go-indent", { clear = true }),
  pattern = "go",
  callback = function()
    vim.opt_local.expandtab = false
    vim.opt_local.tabstop = 4
    vim.opt_local.shiftwidth = 4
  end,
})

-- Return to last edit position when opening files
autocmd("BufReadPost", {
  group = augroup("last-position", { clear = true }),
  callback = function()
    local mark = vim.api.nvim_buf_get_mark(0, '"')
    local line_count = vim.api.nvim_buf_line_count(0)
    if mark[1] > 0 and mark[1] <= line_count then
      pcall(vim.api.nvim_win_set_cursor, 0, mark)
    end
  end,
})

-- Auto-session: save/restore open buffers per project directory
-- Sessions are stored in ~/.local/share/nvim/sessions/
local session_dir = vim.fn.stdpath("data") .. "/sessions/"

local function get_session_path()
  local cwd = vim.fn.getcwd()
  local name = cwd:gsub("/", "%%")
  return session_dir .. name .. ".vim"
end

-- Save session on exit (only if real file buffers are open)
autocmd("VimLeavePre", {
  group = augroup("session-save", { clear = true }),
  callback = function()
    -- Only save if at least one real file buffer exists
    local has_file_buf = false
    for _, buf in ipairs(vim.api.nvim_list_bufs()) do
      if vim.api.nvim_buf_is_loaded(buf)
        and vim.bo[buf].buflisted
        and vim.bo[buf].buftype == ""
        and vim.api.nvim_buf_get_name(buf) ~= "" then
        has_file_buf = true
        break
      end
    end
    if not has_file_buf then return end

    vim.fn.mkdir(session_dir, "p")
    vim.cmd("mksession! " .. vim.fn.fnameescape(get_session_path()))
  end,
})

-- Restore session on start (only when opening nvim with no file arguments)
autocmd("VimEnter", {
  group = augroup("session-restore", { clear = true }),
  nested = true,
  callback = function()
    if vim.fn.argc() > 0 then return end
    local session_file = get_session_path()
    if vim.fn.filereadable(session_file) == 1 then
      vim.cmd("silent! source " .. vim.fn.fnameescape(session_file))
    end
  end,
})
