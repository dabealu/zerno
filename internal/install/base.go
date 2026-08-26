package install

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"zerno/assets"
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
		task.Pacman("update_archlinux_keyring", []string{"archlinux-keyring"}),
		pacstrap(),
		cpuMicrocode(),
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
			if err := steps.Move(binSrc, binDst); err != nil {
				return err
			}

			if !steps.FileExists(confDst) {
				fmt.Printf("moving config %s -> %s\n", confSrc, confDst)
				if err := steps.Move(confSrc, confDst); err != nil {
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

			script := fmt.Sprintf(`iwctl --passphrase '%s' station %s connect '%s'`,
				cfg.WiFiPassword, cfg.NetDevISO, cfg.WiFiSSID)
			if _, err := steps.RunShell(script); err != nil {
				return err
			}

			return steps.WaitForDefaultRoute(20)
		},
	}
}

func partitions() task.Task {
	return task.Task{
		Name: "create_partitions",
		RunFunc: func(cfg *config.Config) error {
			dev := fmt.Sprintf("/dev/%s", cfg.BlockDevice)

			// 1GiB ESP: headroom for current + fallback UKI and future extras
			if _, err := steps.RunCmd("parted", "-s", dev, "mklabel", "gpt"); err != nil {
				return err
			}
			if _, err := steps.RunCmd("parted", "-s", dev,
				"mkpart", "efi-system", "fat32", "1MiB", "1024MiB"); err != nil {
				return err
			}
			if _, err := steps.RunCmd("parted", "-s", dev,
				"mkpart", "rootfs", "ext4", "1024MiB", "100%"); err != nil {
				return err
			}
			if _, err := steps.RunCmd("parted", "-s", dev, "set", "1", "boot", "on"); err != nil {
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
			rootPart := rootPartitionPath(cfg)
			espPart := fmt.Sprintf("/dev/%s%s1", cfg.BlockDevice, cfg.PartNumPrefix)

			if _, err := steps.RunCmd("mkfs.fat", "-F", "32", espPart); err != nil {
				return err
			}
			if _, err := steps.RunCmd("mkfs.ext4", rootPart); err != nil {
				return err
			}
			if _, err := steps.RunCmd("mount", rootPart, "/mnt"); err != nil {
				return err
			}
			if err := os.MkdirAll("/mnt/efi", 0755); err != nil {
				return err
			}
			if _, err := steps.RunCmd("mount", espPart, "/mnt/efi"); err != nil {
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
			pkgs := []string{
				"linux",
				"linux-firmware",
				"base",
				"base-devel",
				"efibootmgr",
				"systemd-ukify",
				"systemd-resolvconf",
				"iwd",
				"impala",
				"python",
				"openssh",
				"dnsutils",
				"curl",
				"git",
				"unzip",
				"neovim",
				"sudo",
				"tmux",
				"sysstat",
				"go",
				"lsof",
				"strace",
				"man",
				"man-db",
				"man-pages",
			}
			args := append([]string{"pacstrap", "/mnt"}, pkgs...)
			_, err := steps.RunCmd(args[0], args[1:]...)
			return err
		},
	}
}

func cpuMicrocode() task.Task {
	return task.Task{
		Name: "install_cpu_microcode",
		RunFunc: func(cfg *config.Config) error {
			out, err := steps.RunCmd("grep", "-m1", "vendor_id", "/proc/cpuinfo")
			if err != nil {
				return err
			}
			var pkg string
			switch {
			case strings.Contains(out, "AuthenticAMD"):
				pkg = "amd-ucode"
			case strings.Contains(out, "GenuineIntel"):
				pkg = "intel-ucode"
			default:
				fmt.Println("unknown CPU vendor, skipping microcode installation")
				return nil
			}
			_, err = steps.RunCmd("pacstrap", "/mnt", pkg)
			return err
		},
	}
}

func setTimezone() task.Task {
	return task.Task{
		Name: "set_timezone",
		RunFunc: func(cfg *config.Config) error {
			zoneinfo := "/usr/share/zoneinfo/" + cfg.Timezone
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "ln", "-sf",
				zoneinfo, "/etc/localtime"); err != nil {
				return err
			}
			_, err := steps.RunCmd("arch-chroot", "/mnt", "hwclock", "--systohc")
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
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "locale-gen"); err != nil {
				return err
			}
			if err := assets.Restore("base/locale.conf", "/mnt/etc/locale.conf"); err != nil {
				return err
			}
			return assets.Restore("base/vconsole.conf", "/mnt/etc/vconsole.conf")
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
			return assets.RestoreTemplate("base/hosts.tpl", "/mnt/etc/hosts", cfg)
		},
	}
}

func user() task.Task {
	return task.Task{
		Name: "create_user",
		RunFunc: func(cfg *config.Config) error {
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "groupadd",
				"-g", strconv.Itoa(cfg.UserGID), cfg.Username); err != nil {
				return err
			}
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "useradd",
				"-m", "-u", strconv.Itoa(cfg.UserID), "-g", strconv.Itoa(cfg.UserGID), cfg.Username); err != nil {
				return err
			}
			if _, err := steps.RunCmd("arch-chroot", "/mnt", "usermod",
				"-aG", "wheel,audio,video,storage", cfg.Username); err != nil {
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
			return config.CheckUEFI()
		},
	}
}

// rootPartitionPath returns the device path of the root partition (partition 2).
func rootPartitionPath(cfg *config.Config) string {
	return fmt.Sprintf("/dev/%s%s%d", cfg.BlockDevice, cfg.PartNumPrefix, cfg.PartNum)
}

// baseKernelCmdline renders /etc/kernel/cmdline for a freshly created root fs.
func baseKernelCmdline(rootUUID string) string {
	return fmt.Sprintf("loglevel=6 root=UUID=%s\n", rootUUID)
}

// hibernationKernelCmdline rewrites cmdline with resume parameters pointing
// at the swapfile (hibernate-to-disk).
func hibernationKernelCmdline(rootUUID, swapUUID, resumeOffset string) string {
	return fmt.Sprintf("loglevel=6 root=UUID=%s resume=UUID=%s resume_offset=%s\n",
		rootUUID, swapUUID, resumeOffset)
}

const (
	loaderConf = "timeout 3\nconsole-mode keep\ndefault arch-linux*\n"

	initramfsHooks = "HOOKS=(base systemd autodetect microcode modconf kms keyboard sd-vconsole block sd-encrypt filesystems fsck)"

	linuxPreset = `# /etc/mkinitcpio.d/linux.preset
PRESETS=('default')
ALL_kver="/boot/vmlinuz-linux"
default_uki="/efi/EFI/Linux/arch-linux.efi"
`

	preserveOldUKIHook = `[Trigger]
Type = File
Operation = Install
Operation = Upgrade
Target = usr/lib/modules/*/vmlinuz

[Action]
Description = Preserving old UKI as fallback...
When = PreTransaction
Exec = /bin/sh -c 'if [ -f /efi/EFI/Linux/arch-linux.efi ]; then cp /efi/EFI/Linux/arch-linux.efi /efi/EFI/Linux/arch-linux-fallback.efi; fi'
`
)

func kernelCmdline() task.Task {
	return task.Task{
		Name: "create_kernel_cmdline",
		RunFunc: func(cfg *config.Config) error {
			rootPart := rootPartitionPath(cfg)

			rootUUID, err := steps.RunCmd("blkid", "-s", "UUID", "-o", "value", rootPart)
			if err != nil {
				return err
			}

			return steps.WriteFile("/mnt/etc/kernel/cmdline",
				baseKernelCmdline(strings.TrimSpace(rootUUID)))
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

			if err := steps.WriteFile("/mnt/efi/loader/loader.conf", loaderConf); err != nil {
				return err
			}

			if err := steps.ReplaceLine("/mnt/etc/mkinitcpio.conf",
				`^HOOKS=.*`, initramfsHooks); err != nil {
				return err
			}

			if err := os.MkdirAll("/mnt/etc/mkinitcpio.d", 0755); err != nil {
				return err
			}
			if err := steps.WriteFile("/mnt/etc/mkinitcpio.d/linux.preset", linuxPreset); err != nil {
				return err
			}

			if err := os.MkdirAll("/mnt/etc/pacman.d/hooks", 0755); err != nil {
				return err
			}
			if err := steps.WriteFile("/mnt/etc/pacman.d/hooks/90-preserve-old-uki.hook",
				preserveOldUKIHook); err != nil {
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
			if !cfg.SecureBoot {
				return nil
			}

			// sbctl is deliberately absent from the main package list -
			// all Secure Boot setup stays encapsulated behind this gate
			if _, err := steps.RunCmd("pacstrap", "/mnt", "sbctl"); err != nil {
				return err
			}
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
