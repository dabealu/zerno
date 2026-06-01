package install

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"zerno/internal/config"
	"zerno/internal/paths"
	"zerno/internal/steps"
	"zerno/internal/task"
)

func Qemu(cfg *config.Config) {
	if err := task.RunTaskList([]task.Task{
		task.RequireUser("root"),
		task.Pacman("install_qemu_packages", []string{"qemu-base", "virt-manager", "dmidecode"}),
		task.Command("add_user_to_libvirt_group", "usermod -a -G libvirt "+cfg.Username),
		task.CopyFile("qemu/qemu0.netdev", "/etc/systemd/network/qemu0.netdev"),
		task.CopyFile("qemu/qemu0.network", "/etc/systemd/network/qemu0.network"),
		task.CopyTemplate("qemu/uplink.network", "/etc/systemd/network/qemu0-uplink.network", cfg),
		task.CopyFile("qemu/bridge.conf", "/etc/qemu/bridge.conf"),
		task.Command("enable_libvirtd_service", "systemctl enable libvirtd"),
		task.Command("start_networkd_and_libvirtd_services", "systemctl restart systemd-networkd libvirtd"),
		task.Command("print_services_status", "systemctl status systemd-networkd libvirtd | grep -E '(.service|Active:)'"),
		task.Info("done, to open gui run `virt-manager`"),
	}, cfg); err != nil {
		log.Fatalf("qemu installation failed: %v", err)
	}
}

func UpdateBin() error {
	if os.Getuid() != 0 {
		return fmt.Errorf("update-bin requires root privileges")
	}

	repoDir := paths.RepoDir(false)
	version := time.Now().Format("02012006-150405")
	script := fmt.Sprintf(`
		cd %s && \
		go fmt ./... && \
		go build -ldflags "-X main.version=%s" -o %s/zerno ./cmd`,
		repoDir, version, repoDir)

	if _, err := steps.RunShell(script); err != nil {
		return err
	}

	fmt.Println("built version:", version)
	binSrc := filepath.Join(repoDir, "zerno")
	binDest := "/usr/local/bin/zerno"
	tmpDest := binDest + ".new"
	if err := steps.CopyRecursive(binSrc, tmpDest); err != nil {
		return err
	}
	if err := os.Rename(tmpDest, binDest); err != nil {
		os.Remove(tmpDest)
		return err
	}
	fmt.Println("done. bin path:", binDest)
	return nil
}

func RepoPull() error {
	homeSrcDir := paths.RepoSrcDir()

	if _, err := os.Stat(homeSrcDir); err == nil {
		cmd := exec.Command("git", "pull")
		cmd.Dir = homeSrcDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pulling: %w", err)
		}
		fmt.Println("updated:", homeSrcDir)
	} else {
		cmd := exec.Command("bash", "-ec", fmt.Sprintf(`mkdir -p %s && git clone %s %s`, filepath.Dir(homeSrcDir), paths.RepoURL, homeSrcDir))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("cloning: %w", err)
		}
		fmt.Println("cloned to:", homeSrcDir)
	}
	return nil
}

func CreateISO() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("root privileges required")
	}

	relengDir := "/usr/share/archiso/configs/releng"
	repoDir := paths.RepoDir(false)
	binPath := filepath.Join(repoDir, "zerno")
	archisoDir := filepath.Join(repoDir, "archiso")
	isoBuildsDir := paths.IsoBuildsDir()
	relengCopyDir := filepath.Join(archisoDir, "releng")

	fmt.Println("building a binary")
	if err := UpdateBin(); err != nil {
		return err
	}

	if err := os.RemoveAll(archisoDir); err != nil {
		return err
	}
	if err := os.MkdirAll(archisoDir, 0755); err != nil {
		return err
	}
	if err := steps.CopyRecursive(relengDir, relengCopyDir); err != nil {
		return err
	}

	relengBinPath := filepath.Join(relengCopyDir, "airootfs/usr/local/bin/zerno")
	if err := steps.CopyFile(binPath, relengBinPath); err != nil {
		return err
	}

	profiledefPath := filepath.Join(relengCopyDir, "profiledef.sh")
	if err := steps.ReplaceLine(profiledefPath, `file_permissions=\(`, "file_permissions=(\n  [\"/usr/local/bin/zerno\"]=\"0:0:755\""); err != nil {
		return err
	}

	fmt.Println("building iso, it may take a while...")
	script := fmt.Sprintf("cd %s && mkarchiso -v -w . -o %s %s", archisoDir, isoBuildsDir, relengCopyDir)
	fmt.Println("running:", script)
	if _, err := steps.RunShell(script); err != nil {
		return err
	}

	uid := 1000
	gid := 1000
	if cfg, err := config.Load(); err == nil {
		if cfg.UserID != 0 {
			uid = cfg.UserID
		}
		if cfg.UserGID != 0 {
			gid = cfg.UserGID
		}
	}
	if err := steps.ChownRecursive(isoBuildsDir, uid, gid); err != nil {
		return err
	}
	if err := os.RemoveAll(archisoDir); err != nil {
		return err
	}

	out, _ := steps.RunCmd("lsblk")
	fmt.Println(out)
	fmt.Println("done, to create installation media run:")
	fmt.Printf("sudo cp %s/archlinux-%s-x86_64.iso /dev/sdX\n", isoBuildsDir, time.Now().Format("2006.01.02"))

	return nil
}

// TODO: organize as a task list
func FormatDevice(devPath, isoPath string) error {
	if os.Getuid() != 0 {
		return fmt.Errorf("boot-dev requires root privileges")
	}
	isoName := filepath.Base(isoPath)
	re := regexp.MustCompile(`archlinux-(\d+\.\d+\.\d+)-x86_64\.iso`)
	matches := re.FindStringSubmatch(isoName)
	if len(matches) < 2 {
		return fmt.Errorf("unable to parse date from iso file name, expected format: archlinux-2022.10.01-x86_64.iso")
	}

	parts := strings.Split(matches[1], ".")
	if len(parts) < 2 {
		return fmt.Errorf("unable to parse date from iso file name")
	}
	isoLabel := fmt.Sprintf("ARCH_%s%s", parts[0], parts[1])

	steps.AskConfirmation(fmt.Sprintf("warning: this will wipe data from %s, continue?", devPath))

	fmt.Println("creating partitions")
	parted := fmt.Sprintf("parted -s %s", devPath)

	for _, script := range []string{
		fmt.Sprintf("%s mklabel gpt", parted),
		fmt.Sprintf("%s mkpart Arch_ISO fat32 1MiB 1024MiB", parted),
		fmt.Sprintf("mkfs.fat -F 32 %s1", devPath),
		fmt.Sprintf("fatlabel %s1 %s", devPath, isoLabel),
	} {
		if _, err := steps.RunShell(script); err != nil {
			return err
		}
	}

	fmt.Printf("copying iso to %s1\n", devPath)
	mntDir := paths.IsoMountDir()
	if err := os.MkdirAll(mntDir, 0755); err != nil {
		return err
	}

	for _, script := range []string{
		fmt.Sprintf("mount %s1 %s", devPath, mntDir),
		fmt.Sprintf("bsdtar -x -f %s -C %s", isoPath, mntDir),
		fmt.Sprintf("umount %s", mntDir),
		fmt.Sprintf("syslinux --directory syslinux --install %s1", devPath),
		fmt.Sprintf("dd bs=440 count=1 conv=notrunc if=/usr/lib/syslinux/bios/gptmbr.bin of=%s", devPath),
		fmt.Sprintf("%s mkpart FlashDrive ext4 1024MiB 100%%", parted),
		fmt.Sprintf("mkfs.ext4 %s2", devPath),
	} {
		if _, err := steps.RunShell(script); err != nil {
			return err
		}
	}

	os.RemoveAll(mntDir)
	fmt.Println("done")
	return nil
}

func ensureMultilib() error {
	content, err := steps.ReadFile("/etc/pacman.conf")
	if err != nil {
		return err
	}

	lines := strings.Split(content, "\n")
	var newLines []string
	skipUntilNextSection := false

	for _, line := range lines {
		multilibRe := regexp.MustCompile(`#?\[multilib\]`)
		if multilibRe.MatchString(line) {
			skipUntilNextSection = true
			continue
		}
		if skipUntilNextSection {
			if strings.HasPrefix(strings.TrimSpace(line), "[") {
				skipUntilNextSection = false
				newLines = append(newLines, line)
			}
			continue
		}
		newLines = append(newLines, line)
	}

	newLines = append(newLines, "[multilib]")
	newLines = append(newLines, "Include = /etc/pacman.d/mirrorlist")

	return steps.WriteFile("/etc/pacman.conf", strings.Join(newLines, "\n"))
}

// TODO: organize as a task list
func InstallSteam(vgaType string) error {
	if os.Getuid() != 0 {
		return fmt.Errorf("steam requires root privileges")
	}
	driverPackages := map[string]string{
		"intel":  "vulkan-intel lib32-vulkan-intel",
		"nvidia": "nvidia-utils lib32-nvidia-utils",
		"amd":    "vulkan-radeon lib32-vulkan-radeon",
	}

	vulkanPackage, ok := driverPackages[vgaType]
	if !ok {
		return fmt.Errorf("unknown vga type: %q, supported values: intel, nvidia, amd", vgaType)
	}

	if err := ensureMultilib(); err != nil {
		return err
	}

	pkgs := []string{
		"ttf-liberation",
		"vulkan-icd-loader",
		"vulkan-tools",
		"lib32-mesa",
		"lib32-systemd",
		"steam",
		vulkanPackage,
	}
	return steps.PacmanPackages(pkgs)
}
