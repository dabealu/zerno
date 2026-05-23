# Steam

## Installation

```sh
zerno steam <vga>    # or: zerno e <vga>
```

This enables multilib and installs Steam + the correct Vulkan driver for your GPU (`intel`, `nvidia`, or `amd`).

If running the command manually, the packages installed are:

| GPU | Packages |
|-----|----------|
| Intel | `vulkan-intel lib32-vulkan-intel` |
| AMD | `vulkan-radeon lib32-vulkan-radeon` |
| NVIDIA | `nvidia-utils lib32-nvidia-utils` |

All modes also install: `ttf-liberation vulkan-icd-loader vulkan-tools lib32-mesa lib32-systemd steam`

> The `lib32-*` Vulkan driver must match your GPU vendor. If pacman prompts for a 32-bit Vulkan driver, do NOT pick `lib32-nvidia-utils` on Intel/AMD.

### System tweaks

Some Proton games need increased `vm.max_map_count`:
```sh
sudo sysctl -w vm.max_map_count=2147483642
```
Make it permanent: `sudo tee /etc/sysctl.d/80-game-compat.conf <<< "vm.max_map_count=2147483642"`

## Proton

Proton (Steam Play) lets you run Windows games on Linux. It's built into Steam.

- Enable in Steam: *Settings → Compatibility → Enable Steam Play*
- "Proton Experimental" is the default since 2024
- Check game compatibility at [Protondb](https://www.protondb.com/)
- To force Proton for a specific game: right-click → *Properties → Compatibility → Force the use of a specific Steam Play compatibility tool*

Proton includes DXVK (DirectX 9/10/11 → Vulkan) and VKD3D-Proton (DirectX 12 → Vulkan), which is why a working Vulkan driver is essential.

## Optimizations

gamemode + gamescope together with a tuned kernel (e.g. CachyOS) provide easy performance gains.
Both packages available in extra repo:
```sh
pacman -S gamescope gamemode lib32-gamemode
```

Example Steam launch options:
```sh
gamemoderun gamescope -w 1280 -h 720 -W 1920 -H 1080 -F fsr -- %command%
```
- `gamemoderun` — sets CPU governor to performance, raises GPU power state, bumps IO/nice priority
- `gamescope` — sandboxed compositor, decouples game rendering from your desktop compositor
- `-w 1280 -h 720` — game renders internally at 720p (saves GPU)
- `-W 1920 -H 1080` — output scaled up to your display resolution
- `-F fsr` — AMD FidelityFX Super Resolution for the upscaling
- `-- %command%` — separator; everything after is the actual game command

## Links

- https://wiki.archlinux.org/title/Steam
- https://wiki.archlinux.org/title/Gamescope
- https://www.protondb.com/
