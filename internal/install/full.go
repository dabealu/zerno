package install

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"zerno/assets"
	"zerno/internal/config"
	"zerno/internal/steps"
	"zerno/internal/task"
)

func Full(cfg *config.Config) {
	if err := task.RunTaskList([]task.Task{
		network(),
		resolved(),
		wifi(),
		globalVars(),
		setupDevTools(),
		swayPackages(),
		task.Info("base desktop installed"),
		swayConfigs(),
		pipewire(),
		swap(),
		hibernation(),
		task.CopyFile("sysctl.d/01-swappiness.conf", "/etc/sysctl.d/01-swappiness.conf"),
		cpuGovernor(),
		task.Command("enable_fstrim_timer", "systemctl enable fstrim.timer"),
		bluetooth(),
		docker(),
		// rustToolchain(),
		userSrcDir(),
		yayAur(),
		aurPackages(),
		task.Command("add_user_to_input_group", "usermod -aG input "+cfg.Username),
		pipewireUser(),
		bashrc(),
		desktopApps(),
		utilsFontsThemes(),
		installUtils(),
		migrateUserConfig(),
		task.Info("installation complete: reboot and run `de`"),
	}, cfg); err != nil {
		log.Fatalf("full installation failed: %v", err)
	}
}

func network() task.Task {
	return task.Task{
		Name: "configure_network",
		RunFunc: func(cfg *config.Config) error {
			if cfg.WiFiEnabled {
				content := `[Match]
Type=wlan

[Network]
DHCP=yes
IgnoreCarrierLoss=3s
`
				if err := steps.WriteFile("/etc/systemd/network/10-wlan.network", content); err != nil {
					return err
				}
			} else {
				content := fmt.Sprintf(`[Match]
Name=%s

[Network]
DHCP=yes
`, cfg.NetDev)
				if err := steps.WriteFile("/etc/systemd/network/0-eth-dhcp.network", content); err != nil {
					return err
				}
			}
			if _, err := steps.RunCmd("systemctl", "enable", "systemd-networkd"); err != nil {
				return err
			}
			_, err := steps.RunCmd("systemctl", "start", "systemd-networkd")
			return err
		},
	}
}

func resolved() task.Task {
	return task.Task{
		Name: "configure_systemd_resolved",
		RunFunc: func(cfg *config.Config) error {
			if err := steps.LineInFile("/etc/resolv.conf", "nameserver 127.0.0.53"); err != nil {
				return err
			}
			if err := os.MkdirAll("/etc/systemd/resolved.conf.d", 0755); err != nil {
				return err
			}
			if err := assets.Restore("files/dns_servers.conf", "/etc/systemd/resolved.conf.d/dns_servers.conf"); err != nil {
				return err
			}
			if _, err := steps.RunCmd("systemctl", "enable", "systemd-resolved"); err != nil {
				return err
			}
			_, err := steps.RunCmd("systemctl", "start", "systemd-resolved")
			return err
		},
	}
}

func SSIDFilename(ssid string) string {
	for _, r := range ssid {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' ||
			r == ' ' || r == '_' || r == '-' {
			continue
		}
		return "=" + hex.EncodeToString([]byte(ssid))
	}
	return ssid
}

func wifi() task.Task {
	return task.Task{
		Name: "configure_iwd_wifi",
		RunFunc: func(cfg *config.Config) error {
			if !cfg.WiFiEnabled {
				return nil
			}

			profilePath := fmt.Sprintf("/var/lib/iwd/%s.psk", SSIDFilename(cfg.WiFiSSID))
			profileContent := fmt.Sprintf(`[Security]
Passphrase=%s

[Settings]
AutoConnect=true
`, cfg.WiFiPassword)
			if err := os.MkdirAll("/var/lib/iwd", 0755); err != nil {
				return err
			}
			if err := steps.WriteFile(profilePath, profileContent); err != nil {
				return err
			}

			if err := assets.Restore("files/iwd-main.conf", "/etc/iwd/main.conf"); err != nil {
				return err
			}

			if _, err := steps.RunCmd("systemctl", "enable", "iwd"); err != nil {
				return err
			}
			if _, err := steps.RunCmd("systemctl", "restart", "iwd"); err != nil {
				return err
			}

			return steps.WaitForDefaultRoute(20)
		},
	}
}

func swayPackages() task.Task {
	return task.Task{
		Name: "install_sway_packages",
		RunFunc: func(cfg *config.Config) error {
			pkgs := []string{
				"sway",
				"swaybg",
				"swaylock",
				"swayidle",
				"waybar",
				"brightnessctl",
				"xorg-xwayland",
				"bemenu-wayland",
				"libnotify",
				"dunst",
				"wl-clipboard",
				"alacritty",
				"ghostty",
			}
			return steps.PacmanPackages(pkgs)
		},
	}
}

func globalVars() task.Task {
	return task.Task{
		Name: "add_global_env_variables",
		RunFunc: func(cfg *config.Config) error {
			for _, v := range []string{"EDITOR=vim", "LIBSEAT_BACKEND=logind"} {
				if err := steps.LineInFile("/etc/environment", v); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func swayConfigs() task.Task {
	return task.Task{
		Name: "create_sway_config_files",
		RunFunc: func(cfg *config.Config) error {
			homeDir := fmt.Sprintf("/home/%s", cfg.Username)
			swayDir := filepath.Join(homeDir, ".config/sway")
			if err := os.MkdirAll(swayDir, 0755); err != nil {
				return err
			}

			swayFiles := map[string]string{
				"conf/alacritty.toml":         swayDir + "/alacritty.toml",
				"conf/config":                 swayDir + "/config",
				"conf/dunstrc":                swayDir + "/dunstrc",
				"conf/libinput-gestures.conf": swayDir + "/libinput-gestures.conf",
				"conf/power-menu.sh":          swayDir + "/power-menu.sh",
				"conf/fav-apps.sh":            swayDir + "/fav-apps.sh",
				"conf/waybar.css":             swayDir + "/waybar.css",
				"conf/waybar.json":            swayDir + "/waybar.json",
				"conf/waybar.sh":              swayDir + "/waybar.sh",
			}
			for src, dst := range swayFiles {
				if err := assets.Restore(src, dst); err != nil {
					return err
				}
			}

			ghosttyDir := filepath.Join(homeDir, ".config/ghostty")
			if err := os.MkdirAll(ghosttyDir, 0755); err != nil {
				return err
			}
			if err := assets.Restore("conf/ghostty", filepath.Join(ghosttyDir, "config")); err != nil {
				return err
			}

			deDst := "/usr/local/bin/de"
			if err := assets.Restore("files/de", deDst); err != nil {
				return err
			}
			if err := os.Chmod(deDst, 0755); err != nil {
				return err
			}
			if err := steps.Symlink(deDst, filepath.Join(homeDir, "de")); err != nil {
				return err
			}

			if err := steps.ChownRecursive(homeDir, cfg.UserID, cfg.UserGID); err != nil {
				return err
			}
			for _, f := range []string{
				".config/sway/waybar.sh",
				".config/sway/power-menu.sh",
				".config/sway/fav-apps.sh",
			} {
				if err := os.Chmod(filepath.Join(homeDir, f), 0755); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func pipewire() task.Task {
	return task.Task{
		Name: "install_pipewire",
		RunFunc: func(cfg *config.Config) error {
			pkgs := []string{
				"pipewire",
				"pipewire-pulse",
				"wireplumber",
				"gst-plugin-pipewire",
				"xdg-desktop-portal-wlr",
			}
			return steps.PacmanPackages(pkgs)
		},
	}
}

func swap() task.Task {
	return task.Task{
		Name: "create_swap_file",
		RunFunc: func(cfg *config.Config) error {
			if steps.FileExists("/swapfile") {
				fmt.Println("/swapfile exists, skipping")
				return nil
			}

			memInfo, err := os.ReadFile("/proc/meminfo")
			if err != nil {
				return err
			}
			var memSizeKB int64
			for _, line := range strings.Split(string(memInfo), "\n") {
				if strings.HasPrefix(line, "MemTotal:") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						parsed, err := strconv.ParseInt(fields[1], 10, 64)
						if err != nil {
							return fmt.Errorf("failed to parse memory size: %w", err)
						}
						memSizeKB = parsed
					}
					break
				}
			}

			if _, err := steps.RunShell(fmt.Sprintf("fallocate -l %dK /swapfile", memSizeKB)); err != nil {
				return err
			}
			if err := os.Chmod("/swapfile", 0600); err != nil {
				return err
			}
			if _, err := steps.RunCmd("mkswap", "/swapfile"); err != nil {
				return err
			}
			if _, err := steps.RunCmd("swapon", "/swapfile"); err != nil {
				return err
			}
			if err := steps.LineInFile("/etc/fstab", "/swapfile none swap defaults 0 0"); err != nil {
				return err
			}
			return nil
		},
	}
}

// Requires swap() to have run first - relies on /swapfile existing
func hibernation() task.Task {
	return task.Task{
		Name: "enable_hibernation_and_suspend",
		RunFunc: func(cfg *config.Config) error {
			rootUUID, err := steps.RunCmd("findmnt", "-no", "UUID", "-T", "/")
			if err != nil {
				return err
			}
			rootUUID = strings.TrimSpace(rootUUID)

			swapDevice, err := steps.RunCmd("findmnt", "-no", "UUID", "-T", "/swapfile")
			if err != nil {
				return err
			}
			swapDevice = strings.TrimSpace(swapDevice)

			out, err := steps.RunCmd("filefrag", "-v", "/swapfile")
			if err != nil {
				return err
			}
			var offset string
			for _, line := range strings.Split(string(out), "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "0:") {
					fields := strings.Fields(line)
					if len(fields) >= 4 {
						offset = strings.TrimSuffix(fields[3], "..")
					}
					break
				}
			}

			cmdline := fmt.Sprintf("loglevel=6 root=UUID=%s resume=UUID=%s resume_offset=%s\n",
				rootUUID, swapDevice, strings.TrimSpace(offset))
			if err := steps.WriteFile("/etc/kernel/cmdline", cmdline); err != nil {
				return err
			}

			if _, err := steps.RunShell("mkinitcpio -p linux"); err != nil {
				return err
			}

			return nil
		},
	}
}

func cpuGovernor() task.Task {
	return task.Task{
		Name: "set_performance_cpu_governor",
		RunFunc: func(cfg *config.Config) error {
			if !steps.FileExists("/sys/devices/system/cpu/cpu0/cpufreq") {
				fmt.Println("cpu doesn't support cpufreq control")
				return nil
			}

			if err := steps.PacmanPackages([]string{"cpupower"}); err != nil {
				return err
			}
			if !steps.FileExists("/etc/default/cpupower") {
				if err := steps.WriteFile("/etc/default/cpupower", "# cpupower defaults\ngovernor='performance'\n"); err != nil {
					return err
				}
			} else {
				if err := steps.ReplaceLine("/etc/default/cpupower", `#governor='ondemand'`, `governor='performance'`); err != nil {
					return err
				}
			}
			if _, err := steps.RunCmd("systemctl", "enable", "cpupower"); err != nil {
				return err
			}
			if _, err := steps.RunCmd("systemctl", "start", "cpupower"); err != nil {
				return err
			}
			data, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor")
			if err != nil {
				return fmt.Errorf("failed to read CPU governor: %w", err)
			}
			fmt.Print(strings.TrimSpace(string(data)))
			return nil
		},
	}
}

func bluetooth() task.Task {
	return task.Task{
		Name: "setup_bluetooth",
		RunFunc: func(cfg *config.Config) error {
			if err := steps.PacmanPackages([]string{
				"bluez",
				"bluez-tools",
				"bluez-utils",
				"blueman",
			}); err != nil {
				return err
			}
			if err := steps.ReplaceLine("/etc/bluetooth/main.conf", `#.*AutoEnable.*`, `AutoEnable = true`); err != nil {
				return err
			}
			if _, err := steps.RunCmd("systemctl", "enable", "bluetooth"); err != nil {
				return err
			}
			_, err := steps.RunCmd("systemctl", "start", "bluetooth")
			return err
		},
	}
}

func docker() task.Task {
	return task.Task{
		Name: "setup_docker",
		RunFunc: func(cfg *config.Config) error {
			if err := steps.PacmanPackages([]string{"docker"}); err != nil {
				return err
			}
			if _, err := steps.RunShell(fmt.Sprintf("usermod -aG docker %s", cfg.Username)); err != nil {
				return err
			}
			if _, err := steps.RunCmd("systemctl", "enable", "docker"); err != nil {
				return err
			}
			_, err := steps.RunCmd("systemctl", "start", "docker")
			return err
		},
	}
}

// func rustToolchain() task.Task {
// 	return task.Task{
// 		Name: "install_rust_toolchain",
// 		RunFunc: func(cfg *config.Config) error {
// 			if _, err := steps.RunCmd("pacman", "-Sy", "--noconfirm", "rustup"); err != nil {
// 				return err
// 			}
// 			script := fmt.Sprintf(`sudo -u %s -- rustup default stable`, cfg.Username)
// 			_, err := steps.RunShell(script)
// 			return err
// 		},
// 	}
// }

func yayAur() task.Task {
	return task.Task{
		Name: "install_yay_aur",
		RunFunc: func(cfg *config.Config) error {
			yayDir := fmt.Sprintf("/home/%s/src/yay", cfg.Username)
			if steps.FileExists(filepath.Join(yayDir, "PKGBUILD")) {
				fmt.Println("yay already cloned, skipping")
				return nil
			}

			if _, err := steps.RunCmd("git", "clone", "https://aur.archlinux.org/yay-git.git", yayDir); err != nil {
				return err
			}
			if err := steps.ChownRecursive(yayDir, cfg.UserID, cfg.UserGID); err != nil {
				return err
			}

			script := fmt.Sprintf("cd %s && sudo -u %s makepkg --noconfirm -si", yayDir, cfg.Username)
			if _, err := steps.RunShell(script); err != nil {
				return err
			}
			return nil
		},
	}
}

func aurPackages() task.Task {
	return task.Task{
		Name: "install_aur_packages",
		RunFunc: func(cfg *config.Config) error {
			pkgs := strings.Join([]string{
				"wdisplays",
				"libinput-gestures",
				"google-chrome",
			}, " ")
			_, err := steps.RunShell(
				fmt.Sprintf("sudo -u %s yay --noconfirm -Sy %s", cfg.Username, pkgs),
			)
			return err
		},
	}
}

func pipewireUser() task.Task {
	return task.Task{
		Name: "start_pipewire",
		RunFunc: func(cfg *config.Config) error {
			userHost := cfg.Username + "@.host"

			if _, err := steps.RunCmd(
				"systemctl", "--user", "-M", userHost, "enable", "pipewire", "pipewire-pulse",
			); err != nil {
				return err
			}

			_, err := steps.RunCmd(
				"systemctl", "--user", "-M", userHost, "start", "pipewire", "pipewire-pulse",
			)
			return err
		},
	}
}

func bashrc() task.Task {
	return task.Task{
		Name: "bashrc_and_user_bin_dir",
		RunFunc: func(cfg *config.Config) error {
			binDir := fmt.Sprintf("/home/%s/bin", cfg.Username)
			if err := os.MkdirAll(binDir, 0755); err != nil {
				return err
			}

			bashrcPath := fmt.Sprintf("/home/%s/.bashrc", cfg.Username)
			if err := assets.Restore("files/bashrc", bashrcPath); err != nil {
				return err
			}

			if err := steps.ChownRecursive(binDir, cfg.UserID, cfg.UserGID); err != nil {
				return err
			}
			return steps.ChownRecursive(bashrcPath, cfg.UserID, cfg.UserGID)
		},
	}
}

func desktopApps() task.Task {
	return task.Task{
		Name: "install_desktop_apps",
		RunFunc: func(cfg *config.Config) error {
			pkgs := []string{
				"evince",
				"telegram-desktop",
				"ristretto",
				"transmission-gtk",
				"vlc",
				"audacious",
				"pavucontrol",
				"thunar",
			}
			return steps.PacmanPackages(pkgs)
		},
	}
}

func setupDevTools() task.Task {
	return task.Task{
		Name: "setup_dev_tools",
		RunFunc: func(cfg *config.Config) error {
			pkgs := []string{
				"ripgrep",
				"fd",
				"fzf",
				"jq",
				"nodejs",
				"npm",
				"simdjson",
				"python-pip",
				"python-pynvim",
				"ttf-jetbrains-mono-nerd",
				"tree-sitter-cli",
				"opencode",
			}
			if err := steps.PacmanPackages(pkgs); err != nil {
				return err
			}

			if _, err := steps.RunShell("npm install -g neovim"); err != nil {
				return err
			}

			nvimDst := fmt.Sprintf("/home/%s/.config/nvim", cfg.Username)
			if err := os.RemoveAll(nvimDst); err != nil {
				return err
			}
			if err := os.MkdirAll(nvimDst, 0755); err != nil {
				return err
			}
			if err := assets.RestoreDir("nvim", nvimDst); err != nil {
				return err
			}

			if err := steps.Symlink("/usr/bin/nvim", "/usr/local/bin/vim"); err != nil {
				return err
			}

			return steps.ChownRecursive(nvimDst, cfg.UserID, cfg.UserGID)
		},
	}
}

func utilsFontsThemes() task.Task {
	return task.Task{
		Name: "install_utilities_fonts_themes",
		RunFunc: func(cfg *config.Config) error {
			pkgs := []string{
				"grim",
				"slurp",
				"ddcutil",
				"nwg-look",
				"syslinux",
				"lshw",
				"pciutils",
				"usbutils",
				"man",
				"man-pages",
				"bash-completion",
				"materia-gtk-theme",
				"papirus-icon-theme",
			}
			if err := steps.PacmanPackages(pkgs); err != nil {
				return err
			}

			homeDir := fmt.Sprintf("/home/%s", cfg.Username)

			gtk3Dir := filepath.Join(homeDir, ".config", "gtk-3.0")
			if err := os.MkdirAll(gtk3Dir, 0755); err != nil {
				return err
			}
			if err := assets.Restore("files/gtk-3.0-settings.ini", filepath.Join(gtk3Dir, "settings.ini")); err != nil {
				return err
			}

			gtk2rc := filepath.Join(homeDir, ".gtkrc-2.0")
			if err := assets.Restore("files/gtk-2.0-gtkrc", gtk2rc); err != nil {
				return err
			}

			if err := steps.ChownRecursive(gtk3Dir, cfg.UserID, cfg.UserGID); err != nil {
				return err
			}
			return steps.ChownRecursive(gtk2rc, cfg.UserID, cfg.UserGID)
		},
	}
}

func installUtils() task.Task {
	return task.Task{
		Name: "install_utilities",
		RunFunc: func(cfg *config.Config) error {
			homeBinDir := fmt.Sprintf("/home/%s/bin", cfg.Username)

			for src, bin := range map[string]string{
				"utilsfs/brightness_control.embed": "brightness-control",
				"utilsfs/translate.embed":          "translate",
			} {
				tmpDir, err := os.MkdirTemp("", "zerno-")
				if err != nil {
					return err
				}
				defer os.RemoveAll(tmpDir)

				srcPath := filepath.Join(tmpDir, strings.ReplaceAll(filepath.Base(src), ".embed", ".go"))
				if err := assets.Restore(src, srcPath); err != nil {
					return err
				}

				dst := filepath.Join(homeBinDir, bin)
				cmd := exec.Command("go", "build", "-o", dst, srcPath)
				cmd.Env = append(os.Environ(), "HOME="+os.Getenv("HOME"))
				out, err := cmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("compile %s: %w\n%s", bin, err, out)
				}
			}

			for _, src := range []string{
				"files/display-disable-laptop.sh",
				"files/display-enable-laptop.sh",
				"files/display-poweroff-external.sh",
			} {
				dst := filepath.Join(homeBinDir, filepath.Base(src))
				if err := assets.Restore(src, dst); err != nil {
					return err
				}
				if err := os.Chmod(dst, 0755); err != nil {
					return err
				}
			}

			return steps.ChownRecursive(homeBinDir, cfg.UserID, cfg.UserGID)
		},
	}
}

func userSrcDir() task.Task {
	return task.Task{
		Name: "create_user_src_dir",
		RunFunc: func(cfg *config.Config) error {
			srcDir := fmt.Sprintf("/home/%s/src", cfg.Username)
			if err := os.MkdirAll(srcDir, 0755); err != nil {
				return err
			}
			return steps.ChownRecursive(srcDir, cfg.UserID, cfg.UserGID)
		},
	}
}

func migrateUserConfig() task.Task {
	return task.Task{
		Name: "migrate_user_config",
		RunFunc: func(cfg *config.Config) error {
			src := "/root/.zerno/parameters.json"
			dstDir := fmt.Sprintf("/home/%s/.zerno", cfg.Username)
			dst := filepath.Join(dstDir, "parameters.json")

			if err := steps.CreateDir(dstDir); err != nil {
				return err
			}
			if err := steps.CopyFile(src, dst); err != nil {
				return err
			}
			return steps.ChownRecursive(dstDir, cfg.UserID, cfg.UserGID)
		},
	}
}
