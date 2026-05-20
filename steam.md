# steam

## amdgpu and vulkan

enable multilib in `/etc/pacman.conf`:
```ini
[multilib]
Include = /etc/pacman.d/mirrorlist
```

install drivers and steam, GPU: `AMD R9 270` - Southern Islands (SI):
```s
pacman -S xf86-video-amdgpu mesa lib32-fontconfig lib32-mesa lib32-systemd vulkan-tools steam
```

add vars to `/etc/environment`
```ini
AMD_VULKAN_ICD=RADV
XDG_RUNTIME_DIR=/run/user/1000
```

load modules, prioritize amdgpu `/etc/mkinitcpio.conf`:
```ini
MODULES=(amdgpu radeon)
```
`mkinitcpio -p linux`

check current driver `lspci -k | less`:
```s
01:00.0 VGA compatible controller: Advanced Micro Devices, Inc. [AMD/ATI] Curacao PRO [Radeon R7 370 / R9 270/370 OEM]
        Subsystem: PC Partner Limited / Sapphire Technology Device e271
        Kernel driver in use: radeon # <<< should be amdgpu instead
        Kernel modules: radeon, amdgpu
```

check `vulkaninfo` output, shouldn't produce error below:
```s
ERROR at /build/vulkan-tools/src/Vulkan-Tools-1.2.199/vulkaninfo/vulkaninfo.h:248:vkEnumeratePhysicalDevices failed with ERROR_INITIALIZATION_FAILED
```

check module parameters:
```s
sudo dmesg | grep -E '(cik|si)_support'
[    1.407151] amdgpu 0000:01:00.0: amdgpu: Use radeon.si_support=0 amdgpu.si_support=1 to override.
```

set module parameters in kernel cmd `/etc/default/grub`:
```ini
GRUB_CMDLINE_LINUX_DEFAULT="... radeon.si_support=0 amdgpu.si_support=1"
```
`sudo grub-mkconfig -o /boot/grub/grub.cfg` and reboot

possible cmd start parameters:
```ini
LD_PRELOAD=/lib64/libSDL2-2.0.so.0 SDL_VIDEODRIVER=wayland DRI_PRIME=1 %command% -vulkan
LD_PRELOAD=/lib64/libSDL2-2.0.so.0 SDL_VIDEODRIVER=x11 DRI_PRIME=1 %command%
```

links:
https://wiki.archlinux.org/title/AMDGPU#Enable_Southern_Islands_(SI)_and_Sea_Islands_(CIK)_support
https://wiki.archlinux.org/title/steam#Installation
https://wiki.archlinux.org/title/Vulkan#AMDGPU_-_ERROR_INITIALIZATION_FAILED_after_vulkaninfo


## desktop environment
in case app requires Xorg or works poorly on wayland.

- install xorg, LXDE and copy xinitrc:
```sh
pacman -Sy xorg xorg-xinit lxde
cp /etc/X11/xinit/xinitrc ~/.xinitrc
```
- edit `~/.xinitrc` and at the bottom of the file comment `twm &` and lines below
- add `exec startlxde`
- start desktop environment with `startx`

## optimizations

worth to try gamemod+gamescope together with cachyos kernel.
this is a minimal effort optimizations bundled in a few easy to use tools.
example command for a Steam command line:

```sh
gamemoderun gamescope -w 1280 -h 720 -W 1920 -H 1080 -F fsr -- %command%
```
- gamemoderun - sets CPU governor to performance, raises GPU power state, bumps IO/nice priority
- gamescope - sandboxed compositor that decouples game rendering from your desktop compositor
- -w 1280 -h 720 - game renders internally at 720p (saves GPU horsepower)
- -W 1920 -H 1080 - output is scaled up to your 1080p display
- -F fsr - uses AMD FidelityFX Super Resolution for the upscaling (sharpens 720p → 1080p)
- -- %command% - the -- separator; everything after is the actual game launch command

