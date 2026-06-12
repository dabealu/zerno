-- Go templates: register as gotmpl filetype (gopls already handles this filetype)
vim.filetype.add({ extension = { gotmpl = "gotmpl" } })

-- blink.cmp: autocompletion
require("blink.cmp").setup({
  keymap = { preset = "super-tab" },
  appearance = {
    nerd_font_variant = "mono",
  },
  completion = {
    documentation = {
      auto_show = true,
      auto_show_delay_ms = 300,
    },
    accept = { auto_brackets = { enabled = true } },
    ghost_text = { enabled = true },
  },
  snippets = {
    preset = "default",
  },
  sources = {
    default = { "lsp", "snippets", "path", "buffer" },
  },
  signature = { enabled = true },
})

-- LSP server configurations (native vim.lsp.config API, Neovim 0.12+)
vim.lsp.config("gopls", {
  settings = {
    gopls = {
      analyses = {
        unusedparams = true,
        shadow = true,
      },
      staticcheck = true,
    },
  },
})

vim.lsp.config("lua_ls", {
  settings = {
    Lua = {
      runtime = {
        version = "LuaJIT",
      },
      diagnostics = {
        globals = { "vim", "Snacks" },
      },
      workspace = {
        library = vim.api.nvim_get_runtime_file("", true),
        checkThirdParty = false,
      },
      telemetry = {
        enable = false,
      },
    },
  },
})

vim.lsp.config("yamlls", {
  settings = {
    yaml = {
      schemas = {
        ["https://json.schemastore.org/github-workflow.json"] = "/.github/workflows/*",
        ["https://json.schemastore.org/kustomization.json"] = "kustomization.yaml",
      },
      validate = true,
    },
  },
})

-- Stubs for servers that need no custom config (Mason picks up from vim.lsp.config)
vim.lsp.config("bashls", {})
vim.lsp.config("jsonls", {})
vim.lsp.config("terraformls", {})
vim.lsp.config("pyright", {})
vim.lsp.config("taplo", {})
vim.lsp.config("marksman", {})

-- Mason: auto-install and auto-enable LSP servers
require("mason").setup()
require("mason-lspconfig").setup({
  ensure_installed = {
    "gopls",
    "lua_ls",
    "yamlls",
    "bashls",
    "jsonls",
    "terraformls",
    "pyright",
    "taplo",
    "marksman",
  },
  automatic_enable = true,
})

-- LSP keymaps (set when an LSP client attaches to a buffer)
vim.api.nvim_create_autocmd("LspAttach", {
  group = vim.api.nvim_create_augroup("lsp-keymaps", { clear = true }),
  callback = function(event)
    local map = function(keys, func, desc)
      vim.keymap.set("n", keys, func, { buffer = event.buf, desc = desc })
    end

    -- Navigation (no leader, high frequency), mirrors keys in gr*
    map("gd", vim.lsp.buf.definition, "Go to definition")
    map("gD", vim.lsp.buf.declaration, "Go to declaration")
    map("K", vim.lsp.buf.hover, "Hover documentation")

    -- LSP actions (leader group)
    vim.keymap.set("n", "<leader>ln", vim.lsp.buf.rename, { buffer = event.buf, desc = "Rename symbol" })
    vim.keymap.set("n", "<leader>la", vim.lsp.buf.code_action, { buffer = event.buf, desc = "Code action" })
    vim.keymap.set("n", "<leader>le", vim.diagnostic.open_float, { buffer = event.buf, desc = "Show error details" })
    vim.keymap.set("n", "<leader>lf", function() vim.lsp.buf.format({ async = false }) end, { buffer = event.buf, desc = "Format buffer" })
  end,
})
