package main

import (
	"fmt"
	"os"

	"zerno/internal/config"
)

var version = "dev"

type cmdDef struct {
	args int
	run  func()
}

var commands = map[string]cmdDef{
	"install-base": {
		run: func() {
			cfg, err := config.LoadOrPrompt()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			runTaskList(installBaseTasks(cfg), cfg)
		},
	},
	"install-full": {
		run: func() {
			cfg, err := config.LoadOrPrompt()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			runTaskList(installFullTasksExt(cfg), cfg)
		},
	},

	"qemu": {
		run: func() {
			cfg, err := config.LoadOrPrompt()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			runTaskList(qemuTasks(cfg), cfg)
		},
	},
	"update-bin": {
		run: func() {
			UpdateBin()
		},
	},
	"build-iso": {
		run: func() {
			CreateISO()
		},
	},
	"boot-dev": {
		args: 2,
		run: func() {
			FormatDevice(os.Args[2], os.Args[3])
		},
	},
	"steam": {
		args: 1,
		run: func() {
			InstallSteam(os.Args[2])
		},
	},
	"version": {
		run: func() {
			fmt.Println(version)
		},
	},
	"repo-pull": {
		run: func() {
			if err := RepoPull(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	},
	"cachyos": {
		run: func() {
			runTaskList(cachyosTasks(), nil)
		},
	},
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		printHelp()
		return
	}

	action := os.Args[1]
	aliases := map[string]string{
		"b": "install-base",
		"i": "install-full",
		"q": "qemu",
		"u": "update-bin",
		"m": "build-iso",
		"f": "boot-dev",
		"e": "steam",
		"v": "version",
		"r": "repo-pull",
		"c": "cachyos",
	}
	if alias, ok := aliases[action]; ok {
		action = alias
	}

	cmd, ok := commands[action]
	args := os.Args[2:]
	if !ok || len(args) < cmd.args {
		printHelp()
		os.Exit(1)
	}

	cmd.run()
}

func printHelp() {
	fmt.Println(`available commands:
  b, install-base       (Phase 1) base system installation (chroot stage)
  i, install-full       (Phase 2) desktop/full installation (after reboot, re-run to sync)
  q, qemu               install and configure qemu/kvm
  c, cachyos            (sudo) enable CachyOS repos and kernel
  u, update-bin         compile new bin from local repo
  m, build-iso          create iso with zerno bin included
  f, boot-dev <dev> <iso>  format device creating storage and boot partitions
  e, steam <vga>        (sudo) install steam, vga: intel, nvidia, amd
  v, version            print version and exit
  r, repo-pull          clone or update repo in ~/src/zerno`)
}
