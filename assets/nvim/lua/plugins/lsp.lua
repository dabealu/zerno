-- Go templates: register as gotmpl filetype (gopls already handles this filetype)
vim.filetype.add({ extension = { gotmpl = "gotmpl" } })

-- Mason: auto-install LSP servers
require("mason").setup()
require("mason-lspconfig").setup({
  automatic_installation = true,
  automatic_enable = true,
})

-- blink.cmp: autocompletion
require("blink.cmp").setup({
  keymap = {
    preset = "default",
    ["<Tab>"] = { "accept", "fallback" },
    ["<CR>"] = { "accept", "fallback" },
    ["<Esc>"] = { "cancel", "fallback" },
    ["<C-e>"] = { "fallback" },
  },
  appearance = {
    nerd_font_variant = "mono",
  },
  completion = {
    documentation = { auto_show = true },
    accept = { auto_brackets = { enabled = true } },
  },
  snippets = {
    preset = "default",
  },
  sources = {
    default = { "lsp", "snippets", "path", "buffer" },
  },
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
      diagnostics = {
        globals = { "vim", "Snacks" },
      },
      workspace = {
        library = vim.api.nvim_get_runtime_file("", true),
        checkThirdParty = false,
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

-- LSP keymaps (set when an LSP client attaches to a buffer)
vim.api.nvim_create_autocmd("LspAttach", {
  group = vim.api.nvim_create_augroup("lsp-keymaps", { clear = true }),
  callback = function(event)
    local map = function(keys, func, desc)
      vim.keymap.set("n", keys, func, { buffer = event.buf, desc = desc })
    end

    -- Navigation (no leader, high frequency)
    map("gd", vim.lsp.buf.definition, "Go to definition")
    vim.keymap.set("n", "gr", vim.lsp.buf.references, { buffer = event.buf, desc = "Go to references", nowait = true })
    map("gi", vim.lsp.buf.implementation, "Go to implementation")
    map("gD", vim.lsp.buf.declaration, "Go to declaration")
    map("gy", vim.lsp.buf.type_definition, "Go to type definition")
    map("K", vim.lsp.buf.hover, "Hover documentation")

    -- LSP actions (leader group)
    vim.keymap.set("n", "<leader>ln", vim.lsp.buf.rename, { buffer = event.buf, desc = "Rename symbol" })
    vim.keymap.set("n", "<leader>la", vim.lsp.buf.code_action, { buffer = event.buf, desc = "Code action" })
    vim.keymap.set("n", "<leader>le", vim.diagnostic.open_float, { buffer = event.buf, desc = "Show error details" })
    vim.keymap.set("n", "<leader>lf", function() vim.lsp.buf.format({ async = false }) end, { buffer = event.buf, desc = "Format buffer" })
  end,
})
