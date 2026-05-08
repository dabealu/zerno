package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"zerno/internal/config"
	"zerno/internal/paths"
	"zerno/internal/steps"
)

func qemuTasks(cfg *config.Config) []Task {
	return []Task{
		requireUser("root"),
		command("install_qemu_packages", "pacman -Sy --noconfirm qemu-base virt-manager dmidecode"),
		command("add_user_to_libvirt_group", "usermod -a -G libvirt "+cfg.Username),
		copyFile("qemu/qemu0.netdev", "/etc/systemd/network/qemu0.netdev"),
		copyFile("qemu/qemu0.network", "/etc/systemd/network/qemu0.network"),
		copyTemplate("qemu/uplink.network", fmt.Sprintf("/etc/systemd/network/qemu0-%s-uplink.network", cfg.NetDev), cfg),
		copyFile("qemu/bridge.conf", "/etc/qemu/bridge.conf"),
		command("enable_libvirtd_service", "systemctl enable libvirtd"),
		command("start_networkd_and_libvirtd_services", "systemctl restart systemd-networkd libvirtd"),
		command("print_services_status", "systemctl status systemd-networkd libvirtd | grep -E '(.service|Active:)'"),
		info("done, to open gui run `virt-manager`"),
	}
}

func UpdateBin() error {
	repoDir := paths.RepoDir(false)

	cmd := exec.Command("bash", "-ec", fmt.Sprintf(`
		VERSION=$(date +%%d%%m%%Y-%%H%%M%%S) && \
		cd %s && \
		go fmt ./... && \
		go build -ldflags "-X main._version=$VERSION" -o %s/zerno ./cmd && \
		echo "Built version: $VERSION"`, repoDir, repoDir))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	binSrc := filepath.Join(repoDir, "zerno")
	binDest := "/usr/local/bin/zerno"
	if _, err := steps.RunShell(fmt.Sprintf("sudo cp -f %s %s", binSrc, binDest)); err != nil {
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

func repoPull() Task {
	return Task{
		Name: "repo_pull",
		RunFunc: func(cfg *config.Config) error {
			return RepoPull()
		},
	}
}

func CreateISO() error {
	if os.Geteuid() != 0 {
		fmt.Println("error: build-iso requires root privileges")
		fmt.Println("run: sudo ./zerno build-iso")
		os.Exit(1)
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

	if _, err := steps.RunShell(fmt.Sprintf("sudo rm -rf %s", archisoDir)); err != nil {
		return err
	}
	if err := os.MkdirAll(archisoDir, 0755); err != nil {
		return err
	}
	if _, err := steps.RunShell(fmt.Sprintf("cp -r %s %s", relengDir, archisoDir)); err != nil {
		return err
	}

	relengBinPath := filepath.Join(relengCopyDir, "airootfs/usr/local/bin/zerno")
	if err := steps.CopyFile(binPath, relengBinPath); err != nil {
		return err
	}

	profiledefPath := filepath.Join(relengCopyDir, "profiledef.sh")
	if err := steps.ReplaceLine(profiledefPath, `file_permissions=\("`, `file_permissions=(\n  ["/usr/local/bin/zerno"]="0:0:755"`); err != nil {
		return err
	}

	fmt.Println("building iso, it may take a while...")
	script := fmt.Sprintf("cd %s && sudo mkarchiso -v -w . -o %s %s", archisoDir, isoBuildsDir, relengCopyDir)
	fmt.Println("running:", script)
	if _, err := steps.RunShell(script); err != nil {
		return err
	}

	userID := "1000"
	userGID := "1000"
	if cfg, err := config.Load(); err == nil {
		if cfg.UserID != "" {
			userID = cfg.UserID
		}
		if cfg.UserGID != "" {
			userGID = cfg.UserGID
		}
	}
	if _, err := steps.RunShell(fmt.Sprintf("sudo chown -R %s:%s %s", userID, userGID, isoBuildsDir)); err != nil {
		return err
	}
	if _, err := steps.RunShell(fmt.Sprintf("sudo rm -rf %s", archisoDir)); err != nil {
		return err
	}

	out, _ := steps.RunCmd("lsblk")
	fmt.Println(out)
	fmt.Println("done, to create installation media run:")
	fmt.Printf("sudo cp %s/archlinux-YYYY.MM.DD-x86_64.iso /dev/sdX\n", isoBuildsDir)

	return nil
}

func FormatDevice(devPath, isoPath string) error {
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
	parted := fmt.Sprintf("sudo parted -s %s", devPath)
	if _, err := steps.RunCmd("bash", "-c", fmt.Sprintf("%s mklabel gpt", parted)); err != nil {
		return err
	}
	if _, err := steps.RunCmd("bash", "-c", fmt.Sprintf("%s mkpart Arch_ISO fat32 1MiB 1024MiB", parted)); err != nil {
		return err
	}
	if _, err := steps.RunCmd("bash", "-c", fmt.Sprintf("sudo mkfs.fat -F 32 %s1", devPath)); err != nil {
		return err
	}
	if _, err := steps.RunCmd("bash", "-c", fmt.Sprintf("sudo fatlabel %s1 %s", devPath, isoLabel)); err != nil {
		return err
	}

	fmt.Printf("copying iso to %s1\n", devPath)
	mntDir := paths.IsoMountDir()
	if err := os.MkdirAll(mntDir, 0755); err != nil {
		return err
	}
	if _, err := steps.RunCmd("bash", "-c", fmt.Sprintf("sudo mount %s1 %s", devPath, mntDir)); err != nil {
		return err
	}
	if _, err := steps.RunCmd("bash", "-c", fmt.Sprintf("sudo bsdtar -x -f %s -C %s", isoPath, mntDir)); err != nil {
		return err
	}
	if _, err := steps.RunCmd("bash", "-c", fmt.Sprintf("sudo umount %s", mntDir)); err != nil {
		return err
	}
	if _, err := steps.RunCmd("bash", "-c", fmt.Sprintf("sudo syslinux --directory syslinux --install %s1", devPath)); err != nil {
		return err
	}
	if _, err := steps.RunCmd("bash", "-c", fmt.Sprintf("sudo dd bs=440 count=1 conv=notrunc if=/usr/lib/syslinux/bios/gptmbr.bin of=%s", devPath)); err != nil {
		return err
	}

	if _, err := steps.RunCmd("bash", "-c", fmt.Sprintf("%s mkpart FlashDrive ext4 1024MiB 100%%", parted)); err != nil {
		return err
	}
	if _, err := steps.RunCmd("bash", "-c", fmt.Sprintf("sudo mkfs.ext4 %s2", devPath)); err != nil {
		return err
	}

	os.RemoveAll(mntDir)
	fmt.Println("done")
	return nil
}

// ensureMultilib removes any existing [multilib] sections (commented or not) and
// appends a clean one at end; pacman uses last definition so this is idempotent.
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

func InstallSteam(vgaType string) error {
	driverPackages := map[string]string{
		"intel":  "vulkan-intel",
		"nvidia": "nvidia-utils",
		"amd":    "amdvlk",
	}

	vulkanPackage, ok := driverPackages[vgaType]
	if !ok {
		return fmt.Errorf("unknown vga type: %s", vgaType)
	}

	if err := ensureMultilib(); err != nil {
		return err
	}

	pkgs := fmt.Sprintf(`%s ttf-liberation vulkan-icd-loader vulkan-tools lib32-mesa lib32-systemd steam`, vulkanPackage)
	_, err := steps.RunCmd("bash", "-c", "pacman -Sy --noconfirm "+pkgs)
	return err
}
