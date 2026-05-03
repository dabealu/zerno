#!/bin/bash
#
# enable-cachyos-repos.sh - CachyOS Repository Setup for Arch Linux
# Usage: sudo ./enable-cachyos-repos.sh

set -euo pipefail

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root (use sudo)"
    exit 1
fi

echo "=== Adding CachyOS repositories ==="

# Download and run CachyOS repo script
cd /tmp
rm -rf cachyos-repo cachyos-repo.tar.xz
curl -fSL https://mirror.cachyos.org/cachyos-repo.tar.xz -o cachyos-repo.tar.xz
tar xf cachyos-repo.tar.xz
cd cachyos-repo
./cachyos-repo.sh

echo "=== Upgrading system with optimized packages ==="
pacman -Syu

echo "=== Installing CachyOS kernel and kernel manager ==="
pacman -S cachyos-kernel-manager linux-cachyos

echo "=== Updating GRUB ==="
/usr/bin/grub-mkconfig -o /boot/grub/grub.cfg

# Set CachyOS kernel as default (works on most GRUB versions)
if ! grep -q "^GRUB_TOP_LEVEL" /etc/default.grub 2>/dev/null; then
    echo 'GRUB_TOP_LEVEL="/boot/vmlinuz-linux-cachyos"' >> /etc/default/grub
else
    sed -i 's|^GRUB_TOP_LEVEL=.*|GRUB_TOP_LEVEL="/boot/vmlinuz-linux-cachyos"|' /etc/default/grub
fi
/usr/bin/grub-mkconfig -o /boot/grub/grub.cfg

echo "=== Done! Reboot to boot CachyOS kernel ==="
echo "Fallback: reboot > GRUB 'Advanced Options' > select 'linux' kernel"