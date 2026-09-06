package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"zerno/assets"
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

// loadConfig loads or prompts for parameters, treating a user-declined
// confirmation as a clean exit.
func loadConfig() *config.Config {
	cfg, err := config.LoadOrPrompt()
	if errors.Is(err, config.ErrAborted) {
		fmt.Println(err)
		os.Exit(0)
	}
	fatalOnErr(err)
	return cfg
}

func printHelp() {
	fmt.Println(`available commands:
  b, install-base          base system installation (chroot)
  i, install-full          full installation and sync configs
  m, build-iso             create iso with zerno bin included
  f, boot-dev <dev> <iso>  format device with storage + boot partitions
  v, version               print version and exit
  r, readme                print embedded README.md to stdout`)
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "b", "install-base":
		install.Base(loadConfig())

	case "i", "install-full":
		install.Full(loadConfig())

	case "m", "build-iso":
		fatalOnErr(install.CreateISO())

	case "f", "boot-dev":
		requireArgCount(4)
		fatalOnErr(install.FormatDevice(os.Args[2], os.Args[3]))

	case "v", "version":
		fmt.Println(version)

	case "r", "readme":
		fmt.Print(assets.Readme())

	default:
		log.Println("unknown command...")
		printHelp()
		os.Exit(1)
	}
}
