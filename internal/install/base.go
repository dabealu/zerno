package install

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"zerno/internal/config"
	"zerno/internal/paths"
	"zerno/internal/steps"
	"zerno/internal/task"
)

func Base(cfg *config.Config) {
	if err := task.RunTaskList([]task.Task{
		task.RequireUser("root"),
		requireUEFI(),
		wifiConnect(),
		partitions(),
		filesystems(),
		kernelCmdline(),
		task.Command("update_archlinux_keyring", "pacman -Sy --noconfirm archlinux-keyring"),
		pacstrap(),
		task.Command("save_fstab", "genfstab -U /mnt >> /mnt/etc/fstab"),
		setTimezone(),
		locales(),
		hostname(),
		user(),
		bootloader(),
		secureBootSign(),
		migrateToChroot(),
		task.Info("reboot and continue installation as root"),
	}, cfg); err != nil {
		// Clean-up mounted block devices
		steps.RunCmd("umount", "-R", "/mnt")
		log.Fatalf("base installation failed: %v", err)
	}
}

func migrateToChroot() task.Task {
	return task.Task{
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

func wifiConnect() task.Task {
	return task.Task{
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

func partitions() task.Task {
	return task.Task{
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

func filesystems() task.Task {
	return task.Task{
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

func pacstrap() task.Task {
	return task.Task{
		Name: "pacstrap_packages",
		RunFunc: func(cfg *config.Config) error {
			pkgs := "linux linux-firmware base base-devel efibootmgr sbctl systemd-ukify systemd-resolvconf wpa_supplicant netplan dbus-python python-rich openssh dnsutils curl git unzip neovim sudo tmux sysstat go lsof strace"
			_, err := steps.RunShell("pacstrap /mnt " + pkgs)
			return err
		},
	}
}

func setTimezone() task.Task {
	return task.Task{
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

func locales() task.Task {
	return task.Task{
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
			if err := task.CopyFile("base/locale.conf", "/mnt/etc/locale.conf").RunFunc(cfg); err != nil {
				return err
			}
			return task.CopyFile("base/vconsole.conf", "/mnt/etc/vconsole.conf").RunFunc(cfg)
		},
	}
}

func hostname() task.Task {
	return task.Task{
		Name: "set_hostname",
		RunFunc: func(cfg *config.Config) error {
			if err := steps.WriteFile("/mnt/etc/hostname", cfg.Hostname); err != nil {
				return err
			}
			return task.CopyTemplate("base/hosts.tpl", "/mnt/etc/hosts", cfg).RunFunc(cfg)
		},
	}
}

func user() task.Task {
	return task.Task{
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

func requireUEFI() task.Task {
	return task.Task{
		Name: "require_uefi",
		RunFunc: func(cfg *config.Config) error {
			if _, err := os.Stat("/sys/firmware/efi"); os.IsNotExist(err) {
				return fmt.Errorf("systemd-boot requires UEFI — /sys/firmware/efi not found")
			}
			return nil
		},
	}
}

func kernelCmdline() task.Task {
	return task.Task{
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

func bootloader() task.Task {
	return task.Task{
		Name: "install_systemd_boot",
		RunFunc: func(cfg *config.Config) error {
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "bootctl",
				"--esp-path=/efi", "install"); err != nil {
				return err
			}

			loaderConf := "timeout 3\nconsole-mode keep\ndefault arch-linux*\n"
			if err := steps.WriteFile("/mnt/efi/loader/loader.conf", loaderConf); err != nil {
				return err
			}

			hooks := "HOOKS=(base systemd autodetect microcode modconf kms keyboard sd-vconsole block sd-encrypt filesystems fsck)"
			if err := steps.ReplaceLine("/mnt/etc/mkinitcpio.conf",
				`^HOOKS=.*`, hooks); err != nil {
				return err
			}

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

			if err := os.MkdirAll("/mnt/efi/EFI/Linux", 0755); err != nil {
				return err
			}

			if _, err := steps.RunCmd("arch-chroot", "/mnt", "mkinitcpio", "-p", "linux"); err != nil {
				return err
			}

			return nil
		},
	}
}

func secureBootSign() task.Task {
	return task.Task{
		Name: "sign_secure_boot",
		RunFunc: func(cfg *config.Config) error {
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "sbctl", "create-keys"); err != nil {
				return err
			}
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "sbctl", "sign", "-s",
				"/efi/EFI/systemd/systemd-bootx64.efi"); err != nil {
				return err
			}
			if err := os.MkdirAll("/mnt/efi/EFI/Linux", 0755); err != nil {
				return err
			}
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "sbctl", "sign", "-s",
				"/efi/EFI/Linux/arch-linux.efi"); err != nil {
				return err
			}

			return nil
		},
	}
}

func baseTasks(cfg *config.Config) []task.Task {
	return []task.Task{
		task.RequireUser("root"),
		requireUEFI(),
		wifiConnect(),
		partitions(),
		filesystems(),
		kernelCmdline(),
		task.Command("update_archlinux_keyring", "pacman -Sy --noconfirm archlinux-keyring"),
		pacstrap(),
		task.Command("save_fstab", "genfstab -U /mnt >> /mnt/etc/fstab"),
		setTimezone(),
		locales(),
		hostname(),
		user(),
		bootloader(),
		secureBootSign(),
		migrateToChroot(),
		task.Info("reboot and continue installation as root"),
	}
}
