package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/prettyletto/leetgo/internal/generator"
)

func writeNeovimDebugBootstrap(dir string, spec *generator.ProblemSpec) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "leetgo_debug.lua")
	content := renderNeovimDebugBootstrap(spec)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func renderNeovimDebugBootstrap(spec *generator.ProblemSpec) string {
	return fmt.Sprintf(`local dap = require("dap")

pcall(function() require("nvim-dap-virtual-text").setup({}) end)

local ui_ok, dapui = pcall(require, "dapui")
if ui_ok then
  dapui.setup({})
  dapui.open({})
end

dap.adapters.go = dap.adapters.go or {
  type = "server",
  port = "${port}",
  executable = {
    command = "dlv",
    args = { "dap", "-l", "127.0.0.1:${port}" },
    detached = true,
  },
}

local config = {
  type = "go",
  name = "Leetgo Debug Case",
  request = "launch",
  mode = "test",
  program = vim.fn.getcwd(),
  args = { "-test.run", "^%s$" },
  outputMode = "remote",
}

vim.g.leetgo_debug_config = config

local function open_ui()
  if ui_ok then
    dapui.open({})
  end
end

local function rerun()
  open_ui()
  dap.run(vim.deepcopy(vim.g.leetgo_debug_config))
end

local function exit_debug()
  dap.terminate()
  if ui_ok then
    dapui.close({})
  end
end

dap.listeners.after.event_terminated["leetgo_keep_ui"] = function()
  vim.defer_fn(function()
    open_ui()
    vim.notify("Debug finished. Press <S-F11> or <leader>dr to rerun, <leader>dx to exit.")
  end, 100)
end
dap.listeners.after.event_exited["leetgo_keep_ui"] = dap.listeners.after.event_terminated["leetgo_keep_ui"]

vim.keymap.set("n", "<S-F11>", rerun, { desc = "Leetgo rerun debug" })
vim.keymap.set("n", "<leader>dr", rerun, { desc = "Leetgo rerun debug" })
vim.keymap.set("n", "<leader>dx", exit_debug, { desc = "Leetgo exit debug" })
vim.keymap.set("n", "<F5>", function()
  if dap.session() then
    dap.continue()
  else
    rerun()
  end
end, { desc = "DAP continue/rerun" })
vim.keymap.set("n", "<F10>", function() dap.step_over() end, { desc = "DAP step over" })
vim.keymap.set("n", "<F11>", function() dap.step_into() end, { desc = "DAP step into" })
vim.keymap.set("n", "<F12>", function() dap.step_out() end, { desc = "DAP step out" })

vim.cmd("normal! zz")
dap.set_breakpoint()
vim.defer_fn(rerun, 100)
`, leetgoDebugTestName)
}

func neovimFuncSearchCommand(spec *generator.ProblemSpec) string {
	return "+/func " + strings.TrimSpace(spec.GoFuncName())
}
