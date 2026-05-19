package main

func cachyosTasks() []Task {
	return []Task{
		command("download_cachyos_repo", "curl -fSL https://mirror.cachyos.org/cachyos-repo.tar.xz -o /tmp/cachyos-repo.tar.xz"),
		command("extract_cachyos_repo", "cd /tmp && rm -rf cachyos-repo && tar xf cachyos-repo.tar.xz"),
		command("run_cachyos_repo_script", "cd /tmp/cachyos-repo && yes | ./cachyos-repo.sh"),
		command("install_cachyos_kernel", "pacman -S --noconfirm cachyos-kernel-manager linux-cachyos"),
		command("preserve_current_kernel_as_fallback",
			"cp /efi/EFI/Linux/arch-linux.efi /efi/EFI/Linux/arch-linux-fallback.efi"),
		command("override_linux_preset_for_cachyos",
			`cat > /etc/mkinitcpio.d/linux.preset << 'EOF'
PRESETS=('default')
ALL_kver="/boot/vmlinuz-linux-cachyos"
default_uki="/efi/EFI/Linux/arch-linux.efi"
EOF
`),
		command("generate_cachyos_uki", "mkinitcpio -p linux"),
		info("done! reboot to activate CachyOS kernel"),
	}
}
