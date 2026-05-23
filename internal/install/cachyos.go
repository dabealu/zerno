package install

import (
	"log"

	"zerno/internal/task"
)

func Cachyos() {
	if err := task.RunTaskList(cachyosTasks(), nil); err != nil {
		log.Fatalf("cachyos installation failed: %v", err)
	}
}

func cachyosTasks() []task.Task {
	return []task.Task{
		task.Command("download_cachyos_repo", "curl -fSL https://mirror.cachyos.org/cachyos-repo.tar.xz -o /tmp/cachyos-repo.tar.xz"),
		task.Command("extract_cachyos_repo", "cd /tmp && rm -rf cachyos-repo && tar xf cachyos-repo.tar.xz"),
		task.Command("run_cachyos_repo_script", "cd /tmp/cachyos-repo && yes | ./cachyos-repo.sh"),
		task.Command("install_cachyos_kernel", "pacman -S --noconfirm cachyos-kernel-manager linux-cachyos"),
		task.Command("preserve_current_kernel_as_fallback",
			"cp /efi/EFI/Linux/arch-linux.efi /efi/EFI/Linux/arch-linux-fallback.efi"),
		task.Command("override_linux_preset_for_cachyos",
			`cat > /etc/mkinitcpio.d/linux.preset << 'EOF'
PRESETS=('default')
ALL_kver="/boot/vmlinuz-linux-cachyos"
default_uki="/efi/EFI/Linux/arch-linux.efi"
EOF
`),
		task.Command("generate_cachyos_uki", "mkinitcpio -p linux"),
		task.Info("done! reboot to activate CachyOS kernel"),
	}
}
