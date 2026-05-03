package main

func cachyosTasks() []Task {
	return []Task{
		command("download_cachyos_repo", "curl -fSL https://mirror.cachyos.org/cachyos-repo.tar.xz -o /tmp/cachyos-repo.tar.xz"),
		command("extract_cachyos_repo", "cd /tmp && rm -rf cachyos-repo && tar xf cachyos-repo.tar.xz"),
		command("run_cachyos_repo_script", "cd /tmp/cachyos-repo && yes | ./cachyos-repo.sh"),
		command("install_cachyos_kernel", "pacman -S --noconfirm cachyos-kernel-manager linux-cachyos"),
		command("set_cachyos_kernel_default", `sh -c 'echo "GRUB_TOP_LEVEL=\"/boot/vmlinuz-linux-cachyos\"" >> /etc/default/grub'`),
		command("update_grub", "grub-mkconfig -o /boot/grub/grub.cfg"),
		info("done! reboot to activate CachyOS kernel"),
	}
}
