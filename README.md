# zerno

Desktop as a Code: automated Arch Linux installation with Sway window manager and Neovim.
Based on: https://github.com/dabealu/arch-sway

## Installation Media

Options:
- zerno build-iso             - create iso with zerno bin included
- zerno boot-dev <dev> <iso>  - format device creating storage and boot partitions

## Quick Start

- Run `sudo zerno install-base` from Arch Linux live environment
- Reboot into new system
- Run `sudo zerno install-full`

## Notes and Troubleshooting 

### Manual wifi configuration
```bash
iwctl station wlan0 connect "SSID"
```
`iwctl station wlan0 get-networks` scans for available networks

### DNS
systemd-resolved is configured with hardcoded public resolvers (`assets/files/dns_servers.conf`);
they override DHCP-provided DNS servers.

### Running on VM
- select `QXL` video device in QEMU, run sway via `WLR_NO_HARDWARE_CURSORS=1 sway`
- archiso environment have sshd and root password access enabled - easy to upload binary and start installation using `scp`/`ssh`

### Flashing black screen during installation
laptop may enter into loop with flashing black screen after selecting install from boot menu.
select `install`, but press `e` instead of `enter` to edit kernel parameters, add `nomodeset` parameter:
```bash
linux /boot/vmlinuz-linux ... nomodeset 
initrd ...
```
press `ctrl+x` to save and load.

ref: https://wiki.archlinux.org/title/Kernel_parameters

### Memory running below rated speed (EXPO/XMP)
RAM often runs at JEDEC baseline instead of the kit's rated profile —
worth ~20% for memory-bound workloads (iGPU gaming, local LLM inference).
check:
```bash
sudo dmidecode -t memory | grep -E 'Speed:|Configured'
```
if `Configured Memory Speed` is lower than the module's rated `Speed`,
the XMP/EXPO profile is disabled — enable it in BIOS (named DOCP on ASUS,
A-XMP on MSI, XMP on Gigabyte). if configured == rated, nothing to do.

### Pipewire
https://wiki.archlinux.org/title/PipeWire
set flag to enable WebRTC in chrome: `chrome://flags/#enable-webrtc-pipewire-capturer`

### Bluetooth pairing
```bash
bluetoothctl
agent KeyboardOnly
default-agent
power on
scan on
pair 00:12:34:56:78:90
connect 00:12:34:56:78:90
```
ref: https://wiki.archlinux.org/title/bluetooth#Pairing

### Screen resolution
run `swaymsg -t get_outputs` to get list of outputs and `man sway-output` for more options.
use `wdisplays` for GUI configuration.

### Appearance
use `nwg-look` to set GTK themes and appearance.

### Connecting android devices via USB
based on: https://wiki.archlinux.org/title/Media_Transfer_Protocol

install dependencies:
`sudo pacman -Sy android-udev android-file-transfer`
restart may be needed.

connect phone, select `File Transfer` (MTP), keep screen unlocked.
mount phone storage:
```bash
mkdir -p ~/mnt
aft-mtp-mount ~/mnt
```

### Keybindings
use `wev` to get key code
```bash
yay -Sy wev
```

### Neovim
Neovim is installed during Phase 1 (base) and configured during Phase 2 (full).
Config is embedded from `assets/nvim/`.
See `vim.md` for detailed documentation, plugins, and keybindings.

### Boot: systemd-boot + UKI
Zerno uses systemd-boot with Unified Kernel Images (UKIs). Kernel cmdline is embedded in the UKI via `/etc/kernel/cmdline`. systemd-boot auto-discovers UKIs in `/efi/EFI/Linux/`. A pacman hook preserves the previous kernel as fallback on upgrades. See `AGENTS.md` for full boot architecture.

### Secure Boot

Secure Boot (SB) is **opt-in**, controlled by the `SecureBoot` parameter
(asked during `install-base`, stored in `~/.zerno/parameters.json`).

- `false` (default): nothing SB-related exists on the system — no sbctl package,
  no keys, no hooks. Zero footprint.
- `true`: zerno creates signing keys and keeps the bootloader and UKI signed
  automatically. Signatures do nothing until SB is activated in firmware — the
  system boots normally either way.

How it works, one line: your private key (on disk) signs boot files at every
rebuild; your public key (enrolled into firmware once) lets the firmware refuse
anything forged at power-on.

#### Enabling (after install)
```bash
# 1. set "SecureBoot": true in ~/.zerno/parameters.json
sudo zerno i            # 2. installs sbctl, creates keys if missing, signs EFI files
sudo sbctl verify       # 3. every file must show ✓ before continuing
# 4. reboot into firmware setup, enter Setup Mode (clear/delete existing SB keys)
sudo sbctl enroll-keys -m   # 5. -m includes Microsoft certs, avoids OptionROM issues
# 6. enable Secure Boot in firmware settings; recommended: also set a BIOS
#    supervisor password - anyone with firmware setup access could just turn SB off
```
`zerno install-full` prints this checklist too.

#### Disabling
- firmware setup → Secure Boot → Disabled. Reversible: signed files stay signed,
  toggling back on later just works.
- if you cleared all keys (Setup Mode) instead: run `sudo sbctl enroll-keys -m`
  again before re-enabling.
- optionally set `"SecureBoot": false` so zerno stops maintaining signatures.
- heads-up: dual-boot Windows BitLocker may demand its recovery key after any
  SB state change.

#### Troubleshooting
- `sbctl status` — current Setup Mode / Secure Boot state.
- `sbctl verify` — ✗ marks unsigned files; fix with `sudo zerno i` or `sudo sbctl sign-all`.
- boot failure right after enabling: switch SB off in firmware, run `sbctl verify`,
  fix ✗ entries, retry. Not a brick — data is untouched by any of this.
- lost/wiped keys (`/var/lib/sbctl` gone): `sudo sbctl create-keys`, then re-run
  `sudo zerno i`; if old keys were already enrolled, replace them via Setup Mode +
  `enroll-keys -m`.
- kernel updates and `install-full` re-runs re-sign automatically; UKIs must only
  ever be regenerated through mkinitcpio.

### TODO
- encrypted volume
- intel integrated graphics: `/etc/modprobe.d/i915.conf`
  ```
  options i915 enable_psr=0 enable_guc=0 enable_fbc=0
  ```
- consider replacing some apps with TUIs:
  - transmission -> aria2
  - thunar       -> yazi
  - ristretto    -> chafa/uberzug++
  - pavucontrol  -> wiremix/pulsemixer
  - evince       -> AUR tdf-git/fancy-cat/just use browser?
  - vlc          -> try for fun: mpv --vo=kitty --hwdec=auto video.mp4
  - audacious    -> create tui wrapper around mpv?

