package install

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"zerno/assets"
	"zerno/internal/config"
	"zerno/internal/steps"
	"zerno/internal/task"
)

func Full(cfg *config.Config) {
	if err := task.RunTaskList([]task.Task{
		task.RequireUser("root"),
		network(),
		resolved(),
		wifiSetup(),
		upgradeSystem(),
		globalVars(),
		setupDevTools(),
		swayPackages(),
		task.Info("base desktop installed"),
		swayConfigs(),
		pipewire(),
		swap(),
		secureBoot(),
		task.CopyFile("sysctl.d/01-swappiness.conf", "/etc/sysctl.d/01-swappiness.conf"),
		splitLockMitigate(),
		cpuGovernor(),
		task.Command("enable_fstrim_timer", "systemctl enable fstrim.timer"),
		bluetooth(),
		docker(),
		userSrcDir(),
		yayAur(),
		aurPackages(),
		voiceToText(),
		task.Run("add_user_to_input_group", "usermod", "-aG", "input", cfg.Username),
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

func upgradeSystem() task.Task {
	return task.Task{
		Name: "upgrade_system",
		RunFunc: func(cfg *config.Config) error {
			_, err := steps.RunCmd("pacman", "-Syu", "--noconfirm")
			return err
		},
	}
}

// wifiDevice returns the wifi interface to manage. NetDevISO holds the
// install-time device name; fall back to the predictable wlan0 name unless it
// looks like a real wireless interface.
func wifiDevice(cfg *config.Config) string {
	dev := cfg.NetDevISO
	if !strings.HasPrefix(dev, "wlan") && !strings.HasPrefix(dev, "wlx") && !strings.HasPrefix(dev, "wlp") {
		return "wlan0"
	}
	return dev
}

// waitForStation blocks until iwctl can address the station, i.e. iwd has come
// up and registered the interface. timeout is in seconds (checked once per
// second).
func waitForStation(dev string, timeout int) error {
	for range timeout {
		if _, err := steps.RunCmd("iwctl", "station", dev, "show"); err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for iwd station %s after %d seconds", dev, timeout)
}

const iwdWaitDropInTemplate = `# iwd can start before the Wi-Fi radio is
# ready and fail its initial autoconnect, leaving the interface down until
# the link is manually cycled. Waiting for the device and forcing it up
# before iwd starts fixes that race (ArchWiki FS#63912).
[Unit]
After=sys-subsystem-net-devices-%s.device
Wants=sys-subsystem-net-devices-%s.device

[Service]
ExecStartPre=ip link set %s up
`

// wifiConnectCmd returns the full iwctl command (binary + argv) that connects
// to ssid on dev, as single unquoted arguments - no shell. The global
// --passphrase option is included only when a password is set so open
// networks can connect too.
func wifiConnectCmd(dev, ssid, pass string) []string {
	cmd := []string{"iwctl"}
	if pass != "" {
		cmd = append(cmd, "--passphrase", pass)
	}
	return append(cmd, "station", dev, "connect", ssid)
}

func wifiSetup() task.Task {
	return task.Task{
		Name: "configure_iwd_wifi",
		RunFunc: func(cfg *config.Config) error {
			if !cfg.WiFiEnabled {
				return nil
			}

			dev := wifiDevice(cfg)
			dropIn := fmt.Sprintf("/etc/systemd/system/iwd.service.d/wait-%s.conf", dev)
			dropInContent := fmt.Sprintf(iwdWaitDropInTemplate, dev, dev, dev)
			if err := os.MkdirAll("/etc/systemd/system/iwd.service.d", 0755); err != nil {
				return err
			}
			if err := steps.WriteFile(dropIn, dropInContent); err != nil {
				return err
			}

			if err := assets.Restore("files/iwd-main.conf", "/etc/iwd/main.conf"); err != nil {
				return err
			}

			if _, err := steps.RunCmd("systemctl", "enable", "iwd"); err != nil {
				return err
			}
			// start, never restart: an already-running daemon is left alone so
			// a working connection survives a re-run of install-full
			if _, err := steps.RunCmd("systemctl", "start", "iwd"); err != nil {
				return err
			}

			// 90s matches systemd's device-unit wait in the drop-in above
			if err := waitForStation(dev, 90); err != nil {
				return err
			}

			// Purge a possibly-stale profile for the configured SSID: a
			// previously unconnectable network would otherwise keep iwd's
			// autoconnect retrying it (and only it) on every boot, leaving
			// wifi dead until a manual connect. Best effort - an absent
			// network is forgotten silently.
			steps.RunCmd("iwctl", "known-networks", cfg.WiFiSSID, "forget")

			// Drive the connection through iwd itself via iwctl.
			cmd := wifiConnectCmd(dev, cfg.WiFiSSID, cfg.WiFiPassword)
			if out, err := steps.RunCmd(cmd[0], cmd[1:]...); err != nil {
				return fmt.Errorf("iwctl connect %q, err: %v, out: %s", cfg.WiFiSSID, err, strings.TrimSpace(out))
			}

			return steps.WaitForDefaultRoute(30)
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
				"python",
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

// swayConfAssets maps embedded asset paths to file names under ~/.config/sway.
var swayConfAssets = map[string]string{
	"conf/alacritty.toml":         "alacritty.toml",
	"conf/config":                 "config",
	"conf/dunstrc":                "dunstrc",
	"conf/libinput-gestures.conf": "libinput-gestures.conf",
	"conf/power-menu.sh":          "power-menu.sh",
	"conf/fav-apps.sh":            "fav-apps.sh",
	"conf/waybar.css":             "waybar.css",
	"conf/waybar.json":            "waybar.json",
	"conf/waybar.sh":              "waybar.sh",
	"conf/waybar-nav.py":          "waybar-nav.py",
}

// swayExecutables get chmod 0755 after restore, relative to the user's home.
var swayExecutables = []string{
	".config/sway/waybar.sh",
	".config/sway/power-menu.sh",
	".config/sway/fav-apps.sh",
	".config/sway/waybar-nav.py",
}

func installSwayFiles(cfg *config.Config, homeDir string) error {
	swayDir := filepath.Join(homeDir, ".config/sway")
	if err := os.MkdirAll(swayDir, 0755); err != nil {
		return err
	}

	for src, name := range swayConfAssets {
		if err := assets.Restore(src, filepath.Join(swayDir, name)); err != nil {
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

	// chown ~/.config wholesale - MkdirAll above can leave the parent
	// root-owned, breaking user services (~/.config is the whole scope
	// ever written here); never walk the whole home directory
	if err := steps.ChownRecursive(filepath.Join(homeDir, ".config"), cfg.UserID, cfg.UserGID); err != nil {
		return err
	}
	for _, f := range swayExecutables {
		if err := os.Chmod(filepath.Join(homeDir, f), 0755); err != nil {
			return err
		}
	}
	return nil
}

func swayConfigs() task.Task {
	return task.Task{
		Name: "create_sway_config_files",
		RunFunc: func(cfg *config.Config) error {
			homeDir := fmt.Sprintf("/home/%s", cfg.Username)
			if err := installSwayFiles(cfg, homeDir); err != nil {
				return err
			}

			deDst := "/usr/local/bin/de"
			if err := assets.Restore("files/de", deDst); err != nil {
				return err
			}
			if err := os.Chmod(deDst, 0755); err != nil {
				return err
			}
			return steps.Symlink(deDst, filepath.Join(homeDir, "de"))
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
				"pipewire-alsa",
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

			if _, err := steps.RunCmd("fallocate", "-l",
				fmt.Sprintf("%dK", memSizeKB), "/swapfile"); err != nil {
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

// secureBoot maintains Secure Boot material when enabled: ensures the sbctl
// package, creates keys if missing and keeps bootloader + UKI signed.
// Disabled means zero footprint - nothing is installed or modified.
func secureBoot() task.Task {
	return task.Task{
		Name: "configure_secure_boot",
		RunFunc: func(cfg *config.Config) error {
			if !cfg.SecureBoot {
				return nil
			}

			if err := steps.PacmanPackages([]string{"sbctl"}); err != nil {
				return err
			}

			if !steps.FileExists("/var/lib/sbctl/keys/db/db.pem") {
				if _, err := steps.RunCmd("sbctl", "create-keys"); err != nil {
					return fmt.Errorf("create keys: %w", err)
				}
			}

			// sign -s is idempotent (skips already-signed files) and keeps
			// paths in the sbctl database for automatic re-signing later
			for _, path := range []string{
				"/efi/EFI/systemd/systemd-bootx64.efi",
				"/efi/EFI/BOOT/BOOTX64.EFI", // firmware fallback path - see secureBootSign
				"/efi/EFI/Linux/arch-linux.efi",
			} {
				if _, err := steps.RunCmd("sbctl", "sign", "-s", path); err != nil {
					return fmt.Errorf("sign %s: %w\nfix hints: run `sbctl status`, inspect README 'Secure Boot' section; broken keys can be recreated via `sbctl create-keys` followed by re-running `zerno install-full`", path, err)
				}
			}

			fmt.Println(`Secure Boot prepared. To activate:
  1. sudo sbctl verify          # every file must show a checkmark
  2. reboot into firmware setup, enter Setup Mode (clear SB keys)
  3. sudo sbctl enroll-keys -m  # -m includes Microsoft certs
  4. enable Secure Boot in firmware settings
see README -> Secure Boot for details and troubleshooting`)
			return nil
		},
	}
}

func cpuGovernor() task.Task {
	return task.Task{
		Name: "set_cpu_governor",
		RunFunc: func(cfg *config.Config) error {
			if !steps.FileExists("/sys/devices/system/cpu/cpu0/cpufreq") {
				fmt.Println("cpu doesn't support cpufreq control")
				return nil
			}

			if err := steps.PacmanPackages([]string{"cpupower"}); err != nil {
				return err
			}

			// Governor comes from parameters.json (prompted at install time,
			// default powersave). cpupower.service reads /etc/default/cpupower-service.conf
			// with UPPERCASE vars; the legacy lowercase /etc/default/cpupower is
			// ignored by the unit, which would silently leave intel_pstate on
			// 'powersave'. Put the setting in the file the unit actually reads.
			gov := cfg.CpuGovernor
			if gov == "" {
				gov = "powersave"
			}
			govLine := "GOVERNOR='" + gov + "'"
			cpupowerConf := "/etc/default/cpupower-service.conf"
			if !steps.FileExists(cpupowerConf) {
				if err := steps.WriteFile(cpupowerConf, govLine+"\n"); err != nil {
					return err
				}
			} else {
				if err := steps.ReplaceLine(cpupowerConf, `(?m)^#?GOVERNOR=.*`, govLine); err != nil {
					return err
				}
				data, err := os.ReadFile(cpupowerConf)
				if err != nil {
					return fmt.Errorf("read cpupower conf: %w", err)
				}
				if !strings.Contains(string(data), "GOVERNOR=") {
					if err := steps.WriteFile(cpupowerConf, string(data)+govLine+"\n"); err != nil {
						return err
					}
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

// splitLockMitigate disables the intel-only split-lock stall (kernel >=5.19
// penalty that wrecks some proton/wine games); amd has no detector, skip.
func splitLockMitigate() task.Task {
	return task.Task{
		Name: "intel_split_lock_mitigate_off",
		RunFunc: func(cfg *config.Config) error {
			cpuInfo, err := os.ReadFile("/proc/cpuinfo")
			if err != nil {
				return err
			}
			if !strings.Contains(string(cpuInfo), "GenuineIntel") {
				fmt.Println("non-intel cpu, skipping split lock sysctl")
				return nil
			}
			return assets.Restore("sysctl.d/30-splitlock.conf", "/etc/sysctl.d/30-splitlock.conf")
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
			if _, err := steps.RunCmd("usermod", "-aG", "docker", cfg.Username); err != nil {
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

			if _, err := steps.RunCmdIn(yayDir, "sudo", "-u", cfg.Username,
				"makepkg", "--noconfirm", "-si"); err != nil {
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
			pkgs := []string{"wdisplays", "libinput-gestures", "google-chrome"}
			argv := append([]string{
				"sudo", "-u", cfg.Username, "yay", "--noconfirm", "-S", "--needed",
			}, pkgs...)
			_, err := steps.RunCmd(argv[0], argv[1:]...)
			return err
		},
	}
}

// voiceToText configures local voice dictation
func voiceToText() task.Task {
	return task.Task{
		Name: "setup_voice_to_text",
		RunFunc: func(cfg *config.Config) error {
			if cfg.VoiceToText == "" {
				return nil
			}

			// wtype simulates keyboard input (type mode)
			if err := steps.PacmanPackages([]string{"wtype"}); err != nil {
				return err
			}

			if _, err := steps.RunCmd("sudo", "-u", cfg.Username, "yay",
				"--noconfirm", "-S", "--needed", "voxtype-bin"); err != nil {
				return err
			}

			voxtypeDir := filepath.Join("/home", cfg.Username, ".config", "voxtype")
			if err := os.MkdirAll(voxtypeDir, 0755); err != nil {
				return err
			}
			if err := assets.RestoreTemplate("conf/voxtype",
				filepath.Join(voxtypeDir, "config.toml"), cfg); err != nil {
				return err
			}
			if err := steps.ChownRecursive(voxtypeDir, cfg.UserID, cfg.UserGID); err != nil {
				return err
			}

			// download model + enable systemd user service, as the target user
			if _, err := steps.RunCmd("sudo", "-u", cfg.Username, "voxtype",
				"setup", "--download", "--no-post-install"); err != nil {
				return err
			}
			if _, err := steps.RunCmd("sudo", "-u", cfg.Username, "voxtype",
				"setup", "systemd"); err != nil {
				return err
			}
			// persist the service so dictation comes up on every login
			if _, err := steps.RunCmd("systemctl", "--user", "-M",
				cfg.Username+"@.host", "enable", "voxtype"); err != nil {
				return err
			}

			// start the daemon now when a user session is already active
			// (re-run/sync case); a fresh install has no session yet and the
			// service comes up on the first login via graphical-session.target
			if _, err := steps.RunCmd("systemctl", "--user", "-M",
				cfg.Username+"@.host", "start", "voxtype"); err != nil {
				fmt.Printf("note: voxtype will start at your next login: %v\n", err)
			}

			return nil
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
				"bash-completion",
				"materia-gtk-theme",
				"papirus-icon-theme",
				// Keep: nerd fonts cover latin/glyphs only, no CJK - without this,
				// Japanese/Chinese/Korean text (e.g. in browser comments) shows as
				// empty boxes. Do not remove.
				"noto-fonts-cjk",
				// Keep: no other font here covers emoji - without this, they
				// render as empty boxes in browsers. Do not remove.
				"noto-fonts-emoji",
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
			systemBinDir := "/usr/local/bin"

			for src, bin := range map[string]string{
				"files/brightness-control.py": "brightness-control",
				"files/translate.py":          "translate",
			} {
				dst := filepath.Join(systemBinDir, bin)
				if err := assets.Restore(src, dst); err != nil {
					return err
				}
				if err := os.Chmod(dst, 0755); err != nil {
					return err
				}
			}

			homeBinDir := fmt.Sprintf("/home/%s/bin", cfg.Username)
			if err := os.MkdirAll(homeBinDir, 0755); err != nil {
				return err
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

			if err := os.MkdirAll(dstDir, 0755); err != nil {
				return err
			}
			// seed the user's copy only once - an existing one holds their
			// edits and must never be overwritten by the /root state
			if !steps.FileExists(dst) {
				if err := steps.CopyFile(src, dst, 0640); err != nil && !os.IsNotExist(err) {
					return err
				}
			}
			return steps.ChownRecursive(dstDir, cfg.UserID, cfg.UserGID)
		},
	}
}
