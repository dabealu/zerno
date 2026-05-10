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
ip link set wlan0 up
wpa_passphrase "SSID" "password" | wpa_supplicant -B -i wlan0 -c /dev/stdin
dhcpcd
```
`wavemon` can help to scan wifi networks

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

### Disable mitigations
this may increase CPU performance, but **potentially dangerous**.
disable hardware vulnerability mitigations by setting `mitigations=off` kernel parameter.

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
use `lxappearance` to set GTK theme and appearance settings.
lxappearance stores config in `~/.gtkrc-2.0`.
more themes: https://wiki.archlinux.org/title/GTK#Themes

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

### TODO
- encrypted volume
- intel integrated graphics
cat /etc/modprobe.d/i915.conf
options i915 enable_psr=0 enable_guc=0 enable_fbc=0
- CPU governor - set permanently to `performance` by default
