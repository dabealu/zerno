# zerno

Automated Arch Linux installation with Sway window manager.

## Quick Start

1. Download latest release or build from source
2. Run `sudo zerno install-base` from Arch Linux live environment
3. Reboot into new system
4. Run `sudo zerno install-full`

## Installation Media

Options:
- **ISO**: Create USB with `zerno boot-dev <dev> <iso>`
- **Arch ISO + binary**: Copy `zerno` to storage partition, boot vanilla Arch ISO, run from there

## Commands

```
b, install-base   base system installation (chroot stage)
i, install-full   desktop/full installation (after reboot)
s, sync           sync configs and desktop settings
q, qemu           install and configure QEMU/KVM
u, update-bin     rebuild binary from source
m, build-iso      create ISO with zerno binary included
f, boot-dev <dev> <iso>  format USB drive with ISO
e, steam <vga>    install Steam (vga: intel, nvidia, amd)
v, version        print version
r, repo-pull      clone or update repo in ~/src/zerno
```

## Configuration

Config stored in `~/.zerno/parameters.json`. Default timezone: `Asia/Singapore`.

## Building from Source

```bash
git clone https://github.com/dabealu/zerno.git ~/src/zerno
cd ~/src/zerno
./build.sh all
```

## Notes

### WiFi Setup (during install-base)
```bash
ip link set wlan0 up
wpa_passphrase "SSID" "password" | wpa_supplicant -B -i wlan0 -c /dev/stdin
dhcpcd
```

### VM Support
- Use QXL video: `WLR_NO_HARDWARE_CURSORS=1 sway`
- SSH available in Arch ISO: `scp zerno root@archiso:/tmp/`

### Troubleshooting

**Flashing black screen**: Add `nomodeset` kernel parameter at boot.

**CPU performance**: Add `mitigations=off` to disable security mitigations.

**Touchpad/Mouse in TTY**: Install `gpm`.

## Features

- Sway window manager
- PipeWire audio
- OpenCode editor
- Steam gaming support
- QEMU/KVM virtualization
- Docker
- Rust toolchain
