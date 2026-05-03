package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"zerno/assets"
	"zerno/internal/config"
	"zerno/internal/steps"
)

func network() Task {
	return Task{
		Name: "configure_network",
		RunFunc: func(cfg *config.Config) error {
			content := fmt.Sprintf(`[Match]
Name=%s
[Network]
DHCP=yes
`, cfg.NetDev)
			if err := steps.WriteFile(fmt.Sprintf("/etc/systemd/network/0-%s-dhcp.network", cfg.NetDev), content); err != nil {
				return err
			}
			if _, err := steps.RunCmd("systemctl", "enable", "systemd-networkd"); err != nil {
				return err
			}
			_, err := steps.RunCmd("systemctl", "start", "systemd-networkd")
			return err
		},
	}
}

func resolved() Task {
	return Task{
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

func netplan() Task {
	return Task{
		Name: "netplan_configuration",
		RunFunc: func(cfg *config.Config) error {
			var assetName, dst string
			if cfg.WiFiEnabled {
				assetName = "files/netplan-wifi-config.yaml"
				dst = "/etc/netplan/wifi-config.yaml"
			} else {
				assetName = "files/netplan-eth-config.yaml"
				dst = "/etc/netplan/eth-config.yaml"
			}

			if err := copyTemplate(assetName, dst, cfg).RunFunc(cfg); err != nil {
				return err
			}
			if err := os.Chmod(dst, 0600); err != nil {
				return err
			}

			if _, err := steps.RunShell("netplan apply"); err != nil {
				return err
			}
			time.Sleep(3 * time.Second)
			out, err := steps.RunCmd("netplan", "get", "all")
			if err != nil {
				return fmt.Errorf("netplan apply verification failed: %w", err)
			}
			fmt.Println(out)
			return nil
		},
	}
}

func swayPackages() Task {
	return Task{
		Name: "install_sway_packages",
		RunFunc: func(cfg *config.Config) error {
			pkgs := "sway swaylock swayidle waybar brightnessctl xorg-xwayland bemenu-wayland libnotify dunst wl-clipboard alacritty ghostty"
			_, err := steps.RunShell("pacman -Sy --noconfirm " + pkgs)
			return err
		},
	}
}

func globalVars() Task {
	return Task{
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

func swayConfigs() Task {
	return Task{
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

			if _, err := steps.RunShell(fmt.Sprintf("chown -R %s:%s %s", cfg.UserID, cfg.UserGID, homeDir)); err != nil {
				return err
			}
			if err := os.Chmod(filepath.Join(homeDir, ".config/sway/waybar.sh"), 0755); err != nil {
				return err
			}

			return nil
		},
	}
}

func pipewire() Task {
	return Task{
		Name: "install_pipewire",
		RunFunc: func(cfg *config.Config) error {
			pkgs := "pipewire pipewire-pulse wireplumber gst-plugin-pipewire xdg-desktop-portal-wlr"
			_, err := steps.RunShell("pacman -Sy --noconfirm " + pkgs)
			return err
		},
	}
}

func swap() Task {
	return Task{
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
						memSizeKB = parsed + 1048576
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

func hibernation() Task {
	return Task{
		Name: "enable_hibernation_and_suspend",
		RunFunc: func(cfg *config.Config) error {
			mkinit, err := steps.ReadFile("/etc/mkinitcpio.conf")
			if err != nil {
				return err
			}
			lines := strings.Split(string(mkinit), "\n")
			for i, line := range lines {
				if strings.HasPrefix(line, "HOOKS=") && !strings.HasSuffix(strings.TrimSpace(line), "resume)") {
					lines[i] = strings.TrimRight(line, ")") + " resume)"
					break
				}
			}
			if err := steps.WriteFile("/etc/mkinitcpio.conf", strings.Join(lines, "\n")); err != nil {
				return err
			}
			if _, err := steps.RunShell("mkinitcpio -p linux"); err != nil {
				return err
			}

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

			grubParams := fmt.Sprintf(`GRUB_CMDLINE_LINUX_DEFAULT="loglevel=3 quiet resume=UUID=%s resume_offset=%s"`, strings.TrimSpace(swapDevice), strings.TrimSpace(offset))
			if err := steps.ReplaceLine("/etc/default/grub", `GRUB_CMDLINE_LINUX_DEFAULT=.*`, grubParams); err != nil {
				return err
			}
			_, err = steps.RunCmd("grub-mkconfig", "-o", "/boot/grub/grub.cfg")
			return err
		},
	}
}

func cpuGovernor() Task {
	return Task{
		Name: "set_performance_cpu_governor",
		RunFunc: func(cfg *config.Config) error {
			if !steps.FileExists("/sys/devices/system/cpu/cpu0/cpufreq") {
				fmt.Println("cpu doesn't support cpufreq control")
				return nil
			}

			if _, err := steps.RunCmd("pacman", "-Sy", "--noconfirm", "cpupower"); err != nil {
				return err
			}
			if err := steps.ReplaceLine("/etc/default/cpupower", `#governor='ondemand'`, `governor='performance'`); err != nil {
				return err
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

func bluetooth() Task {
	return Task{
		Name: "setup_bluetooth",
		RunFunc: func(cfg *config.Config) error {
			if _, err := steps.RunCmd("pacman", "-Sy", "--noconfirm", "bluez", "bluez-tools", "bluez-utils", "blueman"); err != nil {
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

func docker() Task {
	return Task{
		Name: "setup_docker",
		RunFunc: func(cfg *config.Config) error {
			if _, err := steps.RunCmd("pacman", "-Sy", "--noconfirm", "docker"); err != nil {
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

func rustToolchain() Task {
	return Task{
		Name: "install_rust_toolchain",
		RunFunc: func(cfg *config.Config) error {
			if _, err := steps.RunCmd("pacman", "-Sy", "--noconfirm", "rustup"); err != nil {
				return err
			}
			script := fmt.Sprintf(`sudo -u %s -- rustup default stable`, cfg.Username)
			_, err := steps.RunShell(script)
			return err
		},
	}
}

func yayAur() Task {
	return Task{
		Name: "install_yay_aur",
		RunFunc: func(cfg *config.Config) error {
			homeDir := fmt.Sprintf("/home/%s", cfg.Username)
			yayDir := filepath.Join(homeDir, "src/yay-git")
			if steps.FileExists(filepath.Join(yayDir, "PKGBUILD")) {
				fmt.Println("yay already cloned, skipping")
				return nil
			}

			script := fmt.Sprintf(`sudo -u %s -- bash -c 'mkdir -p ~/src && cd ~/src && git clone https://aur.archlinux.org/yay-git.git && cd yay-git && makepkg --noconfirm -si'`, cfg.Username)
			_, err := steps.RunShell(script)
			return err
		},
	}
}

func aurPackages() Task {
	return Task{
		Name: "install_aur_packages",
		RunFunc: func(cfg *config.Config) error {
			pkgs := "google-chrome wdisplays libinput-gestures adwaita-qt5-git adwaita-qt6-git pinta"
			script := fmt.Sprintf(`sudo -u %s -- bash -c 'yes | yay --noconfirm -Sy %s'`, cfg.Username, pkgs)
			_, err := steps.RunShell(script)
			return err
		},
	}
}

func pipewireUser() Task {
	return Task{
		Name: "start_pipewire",
		RunFunc: func(cfg *config.Config) error {
			script := fmt.Sprintf(`systemctl --user -M %s@.host enable pipewire pipewire-pulse && systemctl --user -M %s@.host start pipewire pipewire-pulse`, cfg.Username, cfg.Username)
			_, err := steps.RunShell(script)
			return err
		},
	}
}

func bashrc() Task {
	return Task{
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

			_, err := steps.RunShell(fmt.Sprintf("chown -R %s:%s %s %s", cfg.UserID, cfg.UserGID, binDir, bashrcPath))
			return err
		},
	}
}

func desktopApps() Task {
	return Task{
		Name: "install_desktop_apps",
		RunFunc: func(cfg *config.Config) error {
			pkgs := "code evince libreoffice telegram-desktop ristretto transmission-gtk vlc pavucontrol thunar opencode"
			_, err := steps.RunShell("pacman -Sy --noconfirm " + pkgs)
			return err
		},
	}
}

func configureIDE() Task {
	return Task{
		Name: "configure_editor",
		RunFunc: func(cfg *config.Config) error {
			extensions := []string{
				"golang.go",
				"rust-lang.rust-analyzer",
				"GitHub.github-vscode-theme",
				"PKief.material-icon-theme",
				"ecmel.vscode-html-css",
			}

			for _, ext := range extensions {
				if _, err := steps.RunShell(fmt.Sprintf("sudo -u %s -- code --install-extension %s", cfg.Username, ext)); err != nil {
					return err
				}
			}

			for _, confFile := range []string{"settings.json", "keybindings.json"} {
				dst := fmt.Sprintf("/home/%s/.config/Code - OSS/User/%s", cfg.Username, confFile)
				if err := assets.Restore("files/"+confFile, dst); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func utilsFontsThemes() Task {
	return Task{
		Name: "install_utilities_fonts_themes",
		RunFunc: func(cfg *config.Config) error {
			pkgs := "grim slurp ddcutil lxappearance syslinux lshw pciutils usbutils noto-fonts noto-fonts-cjk noto-fonts-emoji materia-gtk-theme papirus-icon-theme"
			_, err := steps.RunShell("pacman -Sy --noconfirm " + pkgs)
			return err
		},
	}
}

func installUtils() Task {
	return Task{
		Name: "install_utilities",
		RunFunc: func(cfg *config.Config) error {
			homeBinDir := fmt.Sprintf("/home/%s/bin", cfg.Username)
			utils := map[string]string{
				"utilsfs/brightness_control.embed": "brightness-control",
				"utilsfs/translate.embed":          "translate",
			}

			for src, bin := range utils {
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
				if err := cmd.Run(); err != nil {
					return err
				}
			}

			_, err := steps.RunShell(fmt.Sprintf("chown -R %s:%s %s", cfg.UserID, cfg.UserGID, homeBinDir))
			return err
		},
	}
}

func installFullTasksExt(cfg *config.Config) []Task {
	return []Task{
		network(),
		resolved(),
		netplan(),
		swayPackages(),
		info("base desktop installed"),
		globalVars(),
		swayConfigs(),
		pipewire(),
		swap(),
		hibernation(),
		copyFile("sysctl.d/01-swappiness.conf", "/etc/sysctl.d/01-swappiness.conf"),
		cpuGovernor(),
		bluetooth(),
		docker(),
		rustToolchain(),
		yayAur(),
		aurPackages(),
		command("add_user_to_input_group", "usermod -aG input "+cfg.Username),
		pipewireUser(),
		bashrc(),
		desktopApps(),
		configureIDE(),
		utilsFontsThemes(),
		installUtils(),
		userSrcDir(),
		migrateUserConfig(cfg),
		info("installation complete: reboot and run `de`"),
	}
}

func userSrcDir() Task {
	return Task{
		Name: "create_user_src_dir",
		RunFunc: func(cfg *config.Config) error {
			srcDir := fmt.Sprintf("/home/%s/src", cfg.Username)
			if err := os.MkdirAll(srcDir, 0755); err != nil {
				return err
			}
			_, err := steps.RunShell(fmt.Sprintf("chown -R %s:%s %s", cfg.UserID, cfg.UserGID, srcDir))
			return err
		},
	}
}

func migrateUserConfig(cfg *config.Config) Task {
	return Task{
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
			_, err := steps.RunShell(fmt.Sprintf("chown -R %s:%s %s", cfg.UserID, cfg.UserGID, dstDir))
			return err
		},
	}
}
