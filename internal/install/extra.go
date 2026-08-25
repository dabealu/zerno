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

	cmd := exec.Command("./build.sh", "all")
	cmd.Dir = paths.RepoDir(false)
	if err := cmd.Run(); err != nil {
		return err
	}

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

	fmt.Println("done, bin path:", binDest)
	return nil
}

func RepoPull(cfg *config.Config) error {
	homeSrcDir := paths.RepoSrcDir()

	if _, err := os.Stat(homeSrcDir); err == nil {
		cmd := exec.Command("git", "pull")
		cmd.Dir = homeSrcDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pulling: %w", err)
		}
		fmt.Println("updated:", homeSrcDir)
	} else {
		if err := os.MkdirAll(filepath.Dir(homeSrcDir), 0755); err != nil {
			return err
		}
		if _, err := steps.RunCmd("git", "clone", paths.RepoURL, homeSrcDir); err != nil {
			return fmt.Errorf("cloning: %w", err)
		}
		if err := steps.ChownRecursive(homeSrcDir, cfg.UserID, cfg.UserGID); err != nil {
			return err
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
	if _, err := steps.RunCmdIn(archisoDir, "mkarchiso",
		"-v", "-w", ".", "-o", isoBuildsDir, relengCopyDir); err != nil {
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
	var devRe = regexp.MustCompile(`^/dev/[a-zA-Z0-9]+$`)
	if !devRe.MatchString(devPath) {
		return fmt.Errorf("invalid device path %q, expected /dev/<device>", devPath)
	}
	if strings.ContainsAny(isoPath, " \t\n") {
		return fmt.Errorf("iso path must not contain whitespace: %q", isoPath)
	}
	if !steps.FileExists(isoPath) {
		return fmt.Errorf("iso file not found: %s", isoPath)
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

	if !steps.AskConfirmation(fmt.Sprintf("warning: this will wipe data from %s, continue?", devPath)) {
		fmt.Println("aborted")
		return nil
	}

	parted := func(args ...string) error {
		argv := append([]string{"-s", devPath}, args...)
		_, err := steps.RunCmd("parted", argv...)
		return err
	}

	fmt.Println("creating partitions")
	if err := parted("mklabel", "gpt"); err != nil {
		return err
	}
	if err := parted("mkpart", "Arch_ISO", "fat32", "1MiB", "1024MiB"); err != nil {
		return err
	}
	if _, err := steps.RunCmd("mkfs.fat", "-F", "32", devPath+"1"); err != nil {
		return err
	}
	if _, err := steps.RunCmd("fatlabel", devPath+"1", isoLabel); err != nil {
		return err
	}

	fmt.Printf("copying iso to %s1\n", devPath)
	mntDir := paths.IsoMountDir()
	if err := os.MkdirAll(mntDir, 0755); err != nil {
		return err
	}
	if _, err := steps.RunCmd("mount", devPath+"1", mntDir); err != nil {
		return err
	}
	if _, err := steps.RunCmd("bsdtar", "-x", "-f", isoPath, "-C", mntDir); err != nil {
		steps.RunCmd("umount", mntDir)
		os.RemoveAll(mntDir)
		return fmt.Errorf("extract iso: %w", err)
	}
	if _, err := steps.RunCmd("umount", mntDir); err != nil {
		return err
	}
	os.RemoveAll(mntDir)

	for _, argv := range [][]string{
		{"syslinux", "--directory", "syslinux", "--install", devPath + "1"},
		{"dd", "bs=440", "count=1", "conv=notrunc",
			"if=/usr/lib/syslinux/bios/gptmbr.bin", "of=" + devPath},
	} {
		if _, err := steps.RunCmd(argv[0], argv[1:]...); err != nil {
			return err
		}
	}

	if err := parted("mkpart", "FlashDrive", "ext4", "1024MiB", "100%"); err != nil {
		return err
	}
	if _, err := steps.RunCmd("mkfs.ext4", devPath+"2"); err != nil {
		return err
	}

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

// PCI-SIG vendor IDs (static values from the PCI-SIG vendor registry).
const (
	pciVendorIntel  = "0x8086"
	pciVendorAMD    = "0x1002"
	pciVendorNVIDIA = "0x10de"
)

// detectGPUVendor walks <sysFS>/bus/pci/devices and classifies display
// controllers by class prefix 0x03 (0x030000 = VGA, 0x030200 = 3D — NVIDIA
// dGPUs in Optimus laptops enumerate as 3D, not VGA). Vendor IDs are read
// from the same sysfs entries. Discrete AMD wins over Intel iGPU; any NVIDIA
// is reported as unsupported.
func detectGPUVendor(sysFS string) (string, error) {
	devicesDir := filepath.Join(sysFS, "bus", "pci", "devices")
	entries, err := os.ReadDir(devicesDir)
	if err != nil {
		return "", fmt.Errorf("failed to list pci devices: %w", err)
	}

	amd, intel := false, false
	for _, entry := range entries {
		base := filepath.Join(devicesDir, entry.Name())
		class, err := os.ReadFile(filepath.Join(base, "class"))
		if err != nil || !strings.HasPrefix(strings.TrimSpace(string(class)), "0x03") {
			continue
		}
		vendor, err := os.ReadFile(filepath.Join(base, "vendor"))
		if err != nil {
			continue
		}
		switch strings.TrimSpace(string(vendor)) {
		case pciVendorNVIDIA:
			return "", fmt.Errorf("nvidia gpus are not supported yet — see steam.md for manual setup")
		case pciVendorAMD:
			amd = true
		case pciVendorIntel:
			intel = true
		}
	}

	switch {
	case amd:
		return "amd", nil
	case intel:
		return "intel", nil
	default:
		return "", fmt.Errorf("no supported gpu found (looking for amd %s or intel %s display controllers)", pciVendorAMD, pciVendorIntel)
	}
}

// TODO: organize as a task list
func InstallSteam() error {
	if os.Getuid() != 0 {
		return fmt.Errorf("steam requires root privileges")
	}

	vendor, err := detectGPUVendor("/sys")
	if err != nil {
		return err
	}
	fmt.Println("detected gpu vendor:", vendor)

	driverPackages := map[string]string{
		"intel": "vulkan-intel lib32-vulkan-intel",
		"amd":   "vulkan-radeon lib32-vulkan-radeon",
	}

	pkgs := []string{
		"ttf-liberation",
		"vulkan-icd-loader",
		"vulkan-tools",
		"lib32-mesa",
		"lib32-systemd",
		"steam",
		"gamescope",
		driverPackages[vendor],
	}

	if err := ensureMultilib(); err != nil {
		return err
	}

	return steps.PacmanPackages(pkgs)
}
