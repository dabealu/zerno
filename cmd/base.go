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
			if cfg.EFI {
				if _, err := steps.RunShell(fmt.Sprintf("parted -s %s mklabel gpt", dev)); err != nil {
					return err
				}
				if _, err := steps.RunShell(fmt.Sprintf("parted -s %s mkpart efi-system fat32 1MiB 512MiB", dev)); err != nil {
					return err
				}
				if _, err := steps.RunShell(fmt.Sprintf("parted -s %s mkpart rootfs ext4 512MiB 100%%", dev)); err != nil {
					return err
				}
			} else {
				if _, err := steps.RunShell(fmt.Sprintf("parted -s %s mklabel msdos", dev)); err != nil {
					return err
				}
				if _, err := steps.RunShell(fmt.Sprintf("parted -s %s mkpart primary ext4 1MiB 100%%", dev)); err != nil {
					return err
				}
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

			if cfg.EFI {
				if _, err := steps.RunCmd("mkfs.fat", "-F", "32", dev+"1"); err != nil {
					return err
				}
			}
			if _, err := steps.RunCmd("mkfs.ext4", rootPart); err != nil {
				return err
			}
			if _, err := steps.RunCmd("mount", rootPart, "/mnt"); err != nil {
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
			pkgs := "linux linux-firmware base base-devel grub efibootmgr dosfstools os-prober mtools systemd-resolvconf wpa_supplicant netplan dbus-python python-rich openssh dnsutils curl git unzip neovim sudo man man-pages tmux sysstat bash-completion go lsof strace"
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

func grub() Task {
	return Task{
		Name: "install_grub_bootloader",
		RunFunc: func(cfg *config.Config) error {
			if _, err := steps.RunCmd("mkdir", "-p", "/mnt/boot/grub"); err != nil {
				return err
			}

			if cfg.EFI {
				script := fmt.Sprintf(`
					mkdir -p /mnt/boot/EFI && \
					arch-chroot /mnt mount /dev/%s%s1 /boot/EFI && \
					arch-chroot /mnt grub-install --target=x86_64-efi --bootloader-id=grub_uefi --recheck`,
					cfg.BlockDevice, cfg.PartNumPrefix)
				if _, err := steps.RunShell(script); err != nil {
					return err
				}
			} else {
				_, err := steps.RunCmd("arch-chroot", "/mnt", "grub-install", "--recheck", "--target=i386-pc", "/dev/"+cfg.BlockDevice)
				if err != nil {
					return err
				}
			}

			_, err := steps.RunCmd("arch-chroot", "/mnt", "grub-mkconfig", "-o", "/boot/grub/grub.cfg")
			return err
		},
	}
}

func installBaseTasks(cfg *config.Config) []Task {
	return []Task{
		requireUser("root"),
		wifiConnect(),
		partitions(),
		filesystems(),
		command("update_archlinux_keyring", "pacman -Sy --noconfirm archlinux-keyring"),
		pacstrap(),
		command("save_fstab", "genfstab -U /mnt >> /mnt/etc/fstab"),
		setTimezone(),
		locales(),
		hostname(),
		user(),
		grub(),
		migrateToChroot(),
		info("reboot and continue installation as root"),
	}
}
