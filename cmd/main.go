package main

import (
	"fmt"
	"log"
	"os"

	"zerno/internal/config"
	"zerno/internal/install"
)

var version = "dev"

func requireArgCount(n int) {
	if len(os.Args) < n {
		printHelp()
		os.Exit(1)
	}
}

func fatalOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func printHelp() {
	fmt.Println(`available commands:
  b, install-base          base system installation (chroot)
  i, install-full          full installation and sync configs
  q, qemu                  install and configure qemu/kvm
  c, cachyos               enable CachyOS repos and kernel
  u, update-bin            compile new bin from the local repo
  m, build-iso             create iso with zerno bin included
  f, boot-dev <dev> <iso>  format device with storage + boot partitions
  e, steam <vga>           install steam, vga: intel, nvidia, amd
  v, version               print version and exit
  r, repo-pull             clone or update repo in ~/src/zerno`)
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "b", "install-base":
		cfg, err := config.LoadOrPrompt()
		fatalOnErr(err)
		install.Base(cfg)

	case "i", "install-full":
		cfg, err := config.LoadOrPrompt()
		fatalOnErr(err)
		install.Full(cfg)

	case "q", "qemu":
		cfg, err := config.LoadOrPrompt()
		fatalOnErr(err)
		install.Qemu(cfg)

	case "u", "update-bin":
		fatalOnErr(install.UpdateBin())

	case "m", "build-iso":
		fatalOnErr(install.CreateISO())

	case "f", "boot-dev":
		requireArgCount(4)
		fatalOnErr(install.FormatDevice(os.Args[2], os.Args[3]))

	case "e", "steam":
		requireArgCount(3)
		fatalOnErr(install.InstallSteam(os.Args[2]))

	case "v", "version":
		fmt.Println(version)

	case "r", "repo-pull":
		fatalOnErr(install.RepoPull())

	case "c", "cachyos":
		install.Cachyos()

	default:
		log.Println("unknown command...")
		printHelp()
		os.Exit(1)
	}
}
