package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"zerno/internal/config"
	"zerno/internal/paths"
	"zerno/internal/steps"
)

func migrateToChroot() Task {
	return Task{
		Name: "migrate_to_chroot",
		RunFunc: func(cfg *config.Config) error {
			confSrc := paths.ConfDir(false)
			confDst := paths.ConfDir(true)
			binSrc := paths.HostBinPath()
			binDst := "/mnt/usr/local/bin/zerno"

			fmt.Printf("moving binary %s -> %s\n", binSrc, binDst)
			if err := os.MkdirAll("/mnt/usr/local/bin", 0755); err != nil {
				return err
			}
			if _, err := steps.RunShell(fmt.Sprintf("mv %s %s", binSrc, binDst)); err != nil {
				return err
			}

			if !steps.FileExists(confDst) {
				fmt.Printf("moving config %s -> %s\n", confSrc, confDst)
				if _, err := steps.RunShell(fmt.Sprintf("mv %s %s", confSrc, confDst)); err != nil {
					return err
				}
			}

			if err := steps.Symlink(confDst, confSrc); err != nil {
				return err
			}

			return nil
		},
	}
}

func wifiConnect() Task {
	return Task{
		Name: "wifi_connect",
		RunFunc: func(cfg *config.Config) error {
			if !cfg.WiFiEnabled {
				return nil
			}

			out, err := steps.RunCmd("ip", "route", "show", "default")
			if err != nil {
				return err
			}
			if strings.TrimSpace(out) != "" {
				return nil
			}

			script := fmt.Sprintf(`
				ip link set %s up && \
				wpa_supplicant -B -i %s -c <(wpa_passphrase '%s' '%s') && \
				dhcpcd`,
				cfg.NetDevISO, cfg.NetDevISO, cfg.WiFiSSID, cfg.WiFiPassword)
			if _, err := steps.RunShell(script); err != nil {
				return err
			}

			for i := 0; i < 10; i++ {
				time.Sleep(1 * time.Second)
				out, _ := steps.RunCmd("ip", "route", "show", "default")
				if strings.TrimSpace(out) != "" {
					return nil
				}
			}
			return fmt.Errorf("failed to connect to WiFi, timeout reached")
		},
	}
}

func partitions() Task {
	return Task{
		Name: "create_partitions",
		RunFunc: func(cfg *config.Config) error {
			dev := fmt.Sprintf("/dev/%s", cfg.BlockDevice)
			if _, err := steps.RunShell(fmt.Sprintf("parted -s %s mklabel gpt", dev)); err != nil {
				return err
			}
			if _, err := steps.RunShell(fmt.Sprintf("parted -s %s mkpart efi-system fat32 1MiB 512MiB", dev)); err != nil {
				return err
			}
			if _, err := steps.RunShell(fmt.Sprintf("parted -s %s mkpart rootfs ext4 512MiB 100%%", dev)); err != nil {
				return err
			}
			if _, err := steps.RunShell(fmt.Sprintf("parted -s %s set 1 boot on", dev)); err != nil {
				return err
			}
			return nil
		},
	}
}

func filesystems() Task {
	return Task{
		Name: "create_filesystems",
		RunFunc: func(cfg *config.Config) error {
			dev := fmt.Sprintf("/dev/%s%s", cfg.BlockDevice, cfg.PartNumPrefix)
			rootPart := fmt.Sprintf("%s%d", dev, cfg.PartNum)

			if _, err := steps.RunCmd("mkfs.fat", "-F", "32", dev+"1"); err != nil {
				return err
			}
			if _, err := steps.RunCmd("mkfs.ext4", rootPart); err != nil {
				return err
			}
			if _, err := steps.RunCmd("mount", rootPart, "/mnt"); err != nil {
				return err
			}
			if _, err := steps.RunShell("mkdir -p /mnt/efi"); err != nil {
				return err
			}
			if _, err := steps.RunCmd("mount", dev+"1", "/mnt/efi"); err != nil {
				return err
			}
			if _, err := steps.RunCmd("parted", "-s", "/dev/"+cfg.BlockDevice, "print"); err != nil {
				return err
			}
			return nil
		},
	}
}

func pacstrap() Task {
	return Task{
		Name: "pacstrap_packages",
		RunFunc: func(cfg *config.Config) error {
			pkgs := "linux linux-firmware base base-devel efibootmgr dosfstools sbctl systemd-ukify systemd-resolvconf wpa_supplicant netplan dbus-python python-rich openssh dnsutils curl git unzip neovim sudo man man-pages tmux sysstat bash-completion go lsof strace tree-sitter-cli"
			_, err := steps.RunShell("pacstrap /mnt " + pkgs)
			return err
		},
	}
}

func setTimezone() Task {
	return Task{
		Name: "set_timezone",
		RunFunc: func(cfg *config.Config) error {
			script := fmt.Sprintf(`
				arch-chroot /mnt ln -sf /usr/share/zoneinfo/%s /etc/localtime && \
				arch-chroot /mnt hwclock --systohc`, cfg.Timezone)
			_, err := steps.RunShell(script)
			return err
		},
	}
}

func locales() Task {
	return Task{
		Name: "configure_locales",
		RunFunc: func(cfg *config.Config) error {
			if err := steps.ReplaceLine("/mnt/etc/locale.gen", `#.*ru_RU.UTF-8`, `ru_RU.UTF-8`); err != nil {
				return err
			}
			if err := steps.ReplaceLine("/mnt/etc/locale.gen", `#.*en_US.UTF-8`, `en_US.UTF-8`); err != nil {
				return err
			}
			if _, err := steps.RunShell("arch-chroot /mnt locale-gen"); err != nil {
				return err
			}
			if err := copyFile("base/locale.conf", "/mnt/etc/locale.conf").RunFunc(cfg); err != nil {
				return err
			}
			return copyFile("base/vconsole.conf", "/mnt/etc/vconsole.conf").RunFunc(cfg)
		},
	}
}

func hostname() Task {
	return Task{
		Name: "set_hostname",
		RunFunc: func(cfg *config.Config) error {
			if err := steps.WriteFile("/mnt/etc/hostname", cfg.Hostname); err != nil {
				return err
			}
			return copyTemplate("base/hosts.tpl", "/mnt/etc/hosts", cfg).RunFunc(cfg)
		},
	}
}

func user() Task {
	return Task{
		Name: "create_user",
		RunFunc: func(cfg *config.Config) error {
			script := fmt.Sprintf(`
				arch-chroot /mnt groupadd -g %s %s && \
				arch-chroot /mnt useradd -m -u %s -g %s %s && \
				arch-chroot /mnt usermod -aG wheel,audio,video,storage %s`,
				cfg.UserGID, cfg.Username, cfg.UserID, cfg.UserGID, cfg.Username, cfg.Username)
			if _, err := steps.RunShell(script); err != nil {
				return err
			}

			if err := steps.ReplaceLine("/mnt/etc/sudoers", `^#.*%wheel.*NOPASSWD.*$`, `%wheel ALL=(ALL:ALL) NOPASSWD: ALL`); err != nil {
				return err
			}
			if _, err := steps.RunShell(fmt.Sprintf(`arch-chroot /mnt bash -c "echo -e '1\n1' | passwd %s"`, cfg.Username)); err != nil {
				return err
			}
			if _, err := steps.RunShell(`arch-chroot /mnt bash -c "echo -e '1\n1' | passwd root"`); err != nil {
				return err
			}

			fmt.Printf("warning: root and %s passwords are set to '1'\n", cfg.Username)
			return nil
		},
	}
}

func requireUEFI() Task {
	return Task{
		Name: "require_uefi",
		RunFunc: func(cfg *config.Config) error {
			if _, err := os.Stat("/sys/firmware/efi"); os.IsNotExist(err) {
				return fmt.Errorf("systemd-boot requires UEFI — /sys/firmware/efi not found")
			}
			return nil
		},
	}
}

func kernelCmdline() Task {
	return Task{
		Name: "create_kernel_cmdline",
		RunFunc: func(cfg *config.Config) error {
			dev := fmt.Sprintf("/dev/%s%s", cfg.BlockDevice, cfg.PartNumPrefix)
			rootPart := fmt.Sprintf("%s%d", dev, cfg.PartNum)

			rootUUID, err := steps.RunCmd("blkid", "-s", "UUID", "-o", "value", rootPart)
			if err != nil {
				return err
			}

			cmdline := fmt.Sprintf("loglevel=6 root=UUID=%s\n", strings.TrimSpace(rootUUID))
			return steps.WriteFile("/mnt/etc/kernel/cmdline", cmdline)
		},
	}
}

func bootloader() Task {
	return Task{
		Name: "install_systemd_boot",
		RunFunc: func(cfg *config.Config) error {
			// 1. Install systemd-boot to ESP
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "bootctl",
				"--esp-path=/efi", "install"); err != nil {
				return err
			}

			// 2. loader.conf — auto-detect UKIs
			loaderConf := "timeout 3\nconsole-mode keep\ndefault arch-linux*\n"
			if err := steps.WriteFile("/mnt/efi/loader/loader.conf", loaderConf); err != nil {
				return err
			}

			// 3. mkinitcpio HOOKS — systemd-based UKI (from mkinitcpio v40 template)
			hooks := "HOOKS=(base systemd autodetect microcode modconf kms keyboard sd-vconsole block sd-encrypt filesystems fsck)"
			if err := steps.ReplaceLine("/mnt/etc/mkinitcpio.conf",
				`^HOOKS=.*`, hooks); err != nil {
				return err
			}

			// 4. Preset — single UKI. Fallback created on first kernel upgrade.
			preset := `# /etc/mkinitcpio.d/linux.preset
PRESETS=('default')
ALL_kver="/boot/vmlinuz-linux"
default_uki="/efi/EFI/Linux/arch-linux.efi"
`
			if err := os.MkdirAll("/mnt/etc/mkinitcpio.d", 0755); err != nil {
				return err
			}
			if err := steps.WriteFile("/mnt/etc/mkinitcpio.d/linux.preset", preset); err != nil {
				return err
			}

			// 5. Pacman hook: preserve old UKI before kernel upgrade
			hookContent := `[Trigger]
Type = File
Operation = Install
Operation = Upgrade
Target = usr/lib/modules/*/vmlinuz

[Action]
Description = Preserving old UKI as fallback...
When = PreTransaction
Exec = /bin/sh -c 'if [ -f /efi/EFI/Linux/arch-linux.efi ]; then cp /efi/EFI/Linux/arch-linux.efi /efi/EFI/Linux/arch-linux-fallback.efi; fi'
`
			if err := os.MkdirAll("/mnt/etc/pacman.d/hooks", 0755); err != nil {
				return err
			}
			if err := steps.WriteFile("/mnt/etc/pacman.d/hooks/90-preserve-old-uki.hook",
				hookContent); err != nil {
				return err
			}

			// 6. Ensure UKI output directory exists
			if err := os.MkdirAll("/mnt/efi/EFI/Linux", 0755); err != nil {
				return err
			}

			// 7. Generate initial UKI
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "mkinitcpio", "-p", "linux"); err != nil {
				return err
			}

			return nil
		},
	}
}

func secureBootSign() Task {
	return Task{
		Name: "sign_secure_boot",
		RunFunc: func(cfg *config.Config) error {
			// Create Secure Boot keys
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "sbctl", "create-keys"); err != nil {
				return err
			}
			// Sign systemd-boot
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "sbctl", "sign", "-s",
				"/efi/EFI/systemd/systemd-bootx64.efi"); err != nil {
				return err
			}
			// Ensure UKI output directory exists
			if err := os.MkdirAll("/mnt/efi/EFI/Linux", 0755); err != nil {
				return err
			}
			// Sign UKI
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "sbctl", "sign", "-s",
				"/efi/EFI/Linux/arch-linux.efi"); err != nil {
				return err
			}

			return nil
		},
	}
}

func installBaseTasks(cfg *config.Config) []Task {
	return []Task{
		requireUser("root"),
		requireUEFI(),
		wifiConnect(),
		partitions(),
		filesystems(),
		kernelCmdline(),
		command("update_archlinux_keyring", "pacman -Sy --noconfirm archlinux-keyring"),
		pacstrap(),
		command("save_fstab", "genfstab -U /mnt >> /mnt/etc/fstab"),
		setTimezone(),
		locales(),
		hostname(),
		user(),
		bootloader(),
		secureBootSign(),
		migrateToChroot(),
		info("reboot and continue installation as root"),
	}
}
