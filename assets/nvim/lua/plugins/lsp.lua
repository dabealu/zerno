-- Mason: auto-install LSP servers
require("mason").setup()
require("mason-lspconfig").setup({
  ensure_installed = {
    "gopls",
    "terraformls",
    "bashls",
    "yamlls",
    "jsonls",
    "lua_ls",
  },
  automatic_enable = true,
})

-- blink.cmp: autocompletion
require("blink.cmp").setup({
  keymap = {
    preset = "default",
    ["<Tab>"] = { "accept", "fallback" },
    ["<Esc>"] = { "cancel", "fallback" },
    ["<C-e>"] = { "fallback" },
  },
  completion = {
    documentation = { auto_show = true },
  },
  sources = {
    default = { "lsp", "path", "buffer" },
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

-- Enable all configured servers
vim.lsp.enable({
  "gopls",
  "terraformls",
  "bashls",
  "yamlls",
  "jsonls",
  "lua_ls",
})

-- LSP keymaps (set when an LSP client attaches to a buffer)
vim.api.nvim_create_autocmd("LspAttach", {
  group = vim.api.nvim_create_augroup("lsp-keymaps", { clear = true }),
  callback = function(event)
    local bmap = require("config.utils").map
    local buf_opts = { buffer = event.buf }
    local map = function(keys, func, desc)
      vim.keymap.set("n", keys, func, { buffer = event.buf, desc = desc })
    end

    -- Navigation (no leader, high frequency — im-select ensures English)
    map("gd", vim.lsp.buf.definition, "Go to definition")
    map("gr", vim.lsp.buf.references, "Go to references")
    map("gi", vim.lsp.buf.implementation, "Go to implementation")
    map("gD", vim.lsp.buf.declaration, "Go to declaration")
    map("gy", vim.lsp.buf.type_definition, "Go to type definition")
    map("K", vim.lsp.buf.hover, "Hover documentation")

    -- LSP actions (leader group — bilingual)
    bmap("n", "<leader>lr", vim.lsp.buf.rename, "Rename symbol", buf_opts)
    bmap("n", "<leader>la", vim.lsp.buf.code_action, "Code action", buf_opts)
    bmap("n", "<leader>le", vim.diagnostic.open_float, "Show error details", buf_opts)
    bmap("n", "<leader>lf", function() vim.lsp.buf.format({ async = false }) end, "Format buffer", buf_opts)
  end,
})
