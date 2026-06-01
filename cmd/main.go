package main

import (
	"fmt"
	"log"
	"os"

	"zerno/internal/config"
	"zerno/internal/install"
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
			install.Base(cfg)
		},
	},
	"install-full": {
		run: func() {
			cfg, err := config.LoadOrPrompt()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			install.Full(cfg)
		},
	},
	"qemu": {
		run: func() {
			cfg, err := config.LoadOrPrompt()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			install.Qemu(cfg)
		},
	},
	"update-bin": {
		run: func() {
			if err := install.UpdateBin(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	},
	"build-iso": {
		run: func() {
			if err := install.CreateISO(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	},
	"boot-dev": {
		args: 2,
		run: func() {
			if err := install.FormatDevice(os.Args[2], os.Args[3]); err != nil {
				log.Fatal(err)
			}
		},
	},
	"steam": {
		args: 1,
		run: func() {
			if err := install.InstallSteam(os.Args[2]); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	},
	"version": {
		run: func() {
			fmt.Println(version)
		},
	},
	"repo-pull": {
		run: func() {
			if err := install.RepoPull(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	},
	"cachyos": {
		run: func() {
			install.Cachyos()
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
  b, install-base       base system installation (chroot stage)
  i, install-full       desktop/full installation (after reboot, re-run to sync)
  q, qemu               install and configure qemu/kvm
  c, cachyos            enable CachyOS repos and kernel
  u, update-bin         compile new bin from local repo
  m, build-iso          create iso with zerno bin included
  f, boot-dev <dev> <iso>  format device creating storage and boot partitions
  e, steam <vga>        install steam, vga: intel, nvidia, amd
  v, version            print version and exit
  r, repo-pull          clone or update repo in ~/src/zerno`)
}
