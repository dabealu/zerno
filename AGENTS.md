# Zerno Agent Guidelines

## Project Overview

Zerno is a Go tool for automated Arch Linux installation with Sway window manager.

## Key Features

- **Standalone binary** - all configs/templates embedded via `//go:embed`
- **Two-phase installation** - `install-base` (Phase 1, chroot) + `install-full` (Phase 2, after reboot)
- **Re-run `install-full` to sync** - re-running Phase 2 syncs configs and packages
- **No repo required during install** - binary is self-contained

## Key Packages

### internal/paths

**Always use this package for path resolution.** Don't hardcode paths.

### internal/assets

Embedded assets (configs, templates). Files in `assets/` are embedded at compile time.
Templates use Go's `text/template` and receive `*config.Config` for substitution:

### internal/steps

Contains function to perform operations with files and run commands

- `PacmanPackages()` runs `pacman -S --needed --noconfirm` (**no `-Sy`/`-u`**) -
  package installs are never partial upgrades (unsupported in Arch; a `-Sy`
  install once pulled a freshly-rebuilt `nodejs` that required a newer `ada`,
  breaking node mid-install). Freshness is guaranteed by the `upgrade_system`
  task at the start of `install-full` running `pacman -Syu --noconfirm`, so all
  later installs operate on a synced, consistent system. Flag-driven installs
  (e.g. `steam()`, `qemu()`) also assume a recently synced system.
- Exception: `update_archlinux_keyring` in `install-base` runs `pacman -Sy`
  directly - it runs on the disposable live ArchISO (nothing installed, no
  partial-upgrade hazard), where the frozen ISO database would otherwise pin
  an archlinux-keyring version the mirrors no longer carry and fail outright.

### internal/config

Config struct and loading. Saved to `parameters.json` in the user's `.zerno/`
dir (`~/.zerno/parameters.json` when run as root via sudo; `/root/.zerno/parameters.json`
when running in a direct root shell like the live ArchISO, so Phase 1 writes it
onto the installed disk for Phase 2 to pick up).

## Code Conventions

### Task Functions

Each task is a function returning a `Task` struct with cfg parameter:

```go
func myTask(cfg *config.Config) Task {
    return Task{
        Name: "task_name",
        RunFunc: func(cfg *config.Config) error {
            // implementation
            return nil
        },
    }
}
```

Tasks are executed via `task.RunTaskList(tasks, cfg)`.

### Adding Config Files and Templates

1. Add file/template to `assets/` directory
2. Use `assets.Restore("path/in/assets", "/destination/path")` or `assets.RestoreTemplate("path/in/assets", "/destination/path", cfg)`
3. For directories (like `nvim/`), use `assets.RestoreDir("path/in/assets", "/destination/path")`

### Utility Scripts

System utilities live in `assets/files/*.py` (python, stdlib only) and are
restored to `/usr/local/bin/<name>` (extensionless) with exec bit by `installUtils()`.
No compilation step. Desktop scripts (`.sh`, `waybar-nav.py`) go to `~/.config/sway/`
via `swayConfigs()` + `swayExecutables` chmod list.

## Commands

```bash
./build.sh               # fmt, vet, test, build, and install to /usr/local/bin/zerno
./build.sh --no-install  # fmt, vet, test, build only (outputs ./zerno in repo root)
```

`build.sh` has no other subcommands - it always runs fmt+vet+test+build. With no
args it then installs the fresh binary to `/usr/local/bin/zerno` (via sudo when
not root). `--no-install` just builds so callers (e.g. `build-iso`,
`CreateISO`) can copy `./zerno` wherever they need it.

## Available Commands

| Command | Alias | Description |
|---------|-------|-------------|
| install-base | b | (Phase 1) Run base system installation (chroot stage) |
| install-full | i | (Phase 2) Desktop/full installation (after reboot, re-run to sync) |
| build-iso | m | Create iso with zerno bin included |
| boot-dev | f | Format device creating storage and boot partitions |
| version | v | Print version and exit |
| readme | r | Print embedded README.md to stdout |

Steam and QEMU are no longer separate commands - they are `Steam`/`Qemu` bool
config params (default false), installed by the `steam()`/`qemu()` tasks during
`install-full` when enabled.

## Testing

- Integration tests use real filesystem in temp dirs
- Set `HOME` env var for tests needing config directories

### Open: command-runner test seam (discuss before implementing)

The suite covers file ops, config, assets — but not **what commands zerno runs**
(pacman args, iwctl argv). That blind spot let 4 bugs through (partial `-Sy`
upgrade, iwctl SSID quoting, upgrade_system ordering, SUDO_USER resolve).
Planned: a steps-level `runCmd` seam + recorder tests asserting e.g.
`pacman -S --needed`, `pacman -Syu`, and iwctl `connect <ssid>` as one unquoted
argv element. Do NOT implement without discussing design first.

## Neovim Config

Neovim configuration is in `assets/nvim/` (embedded, deployed via `install-full`).
See [vim.md](vim.md) for plugin docs, keybindings, and [vim-cheatsheet.md](vim-cheatsheet.md) for a quick reference.

## Boot Architecture

### Bootloader: systemd-boot + UKIs

- **systemd-boot** — simple UEFI boot manager, installed to ESP at `/efi/`
- **UKIs** (Unified Kernel Images) — single `.efi` file containing kernel + initramfs + cmdline, generated by mkinitcpio (offloaded to **ukify** when `systemd-ukify` is installed, which is always)
- **ESP layout:**
  ```
  /efi/EFI/systemd/               ← systemd-boot EFI application
  /efi/EFI/Linux/arch-linux.efi   ← current kernel (UKI)
  /efi/EFI/Linux/arch-linux-fallback.efi  ← previous kernel (preserved on upgrade)
  ```
- **Auto-discovery** — systemd-boot scans `/efi/EFI/Linux/` and creates a menu entry for every `.efi` file found. Drop a file, it appears. No config needed.
- **`/boot/`** stays on root partition — contains the kernel binary (`vmlinuz-linux`) as a **build input** for mkinitcpio, NOT a bootloader directory
- **MBR/BIOS support dropped** — systemd-boot requires UEFI

### Boot flow
```
Power on → firmware → systemd-boot on ESP → menu → 
  loads UKI (.efi with kernel + initramfs + cmdline) → boots system
```

### Kernel cmdline
- Stored in `/etc/kernel/cmdline` (not `/etc/default/grub`)
- Baked into the UKI by mkinitcpio's `systemd` hook
- Base install writes: `loglevel=6 root=UUID=...` (Phase 1)
- Phase 2 `hibernation()` only rebuilds the UKI — no resume params needed; systemd writes `HibernateLocation` EFI variable automatically on `systemctl hibernate`
- Future LUKS: add `rd.luks.name=...` from a separate task

### mkinitcpio hooks
```
HOOKS=(base systemd autodetect microcode modconf kms keyboard sd-vconsole block sd-encrypt filesystems fsck)
```
- `systemd` replaces `udev` — provides systemd in initramfs, native resume handling
- No separate `resume` hook needed (systemd handles it via `systemd-hibernate-resume.service`)
- `microcode` loads CPU microcode updates early (no-op if intel-ucode/amd-ucode not installed)
- `sd-vconsole` manages console font/keymap via systemd (reads `/etc/vconsole.conf`)
- `sd-encrypt` for future LUKS support (no-op when no encrypted volumes configured)
- **Source of truth**: the upstream [mkinitcpio.conf](https://github.com/archlinux/mkinitcpio/blob/master/mkinitcpio.conf) contains commented examples; our list is the "systemd + encrypted root" example verbatim. When Arch changes hooks, it'll be front-page news. Review the constant `initramfsHooks` in `base.go` against upstream source once in a while.

### UKI fallback strategy
- Initial install: single UKI (`arch-linux.efi`)
- Pacman hook (`00-preserve-old-uki.hook`) runs **PreTransaction** on kernel upgrades
  (the `00-` prefix keeps it ahead of mkinitcpio's own `60-mkinitcpio-remove.hook`,
  which deletes the old UKI at PreTransaction — same-When hooks run in filename
  order, a later-named preserve hook would find nothing to copy):
  copies `arch-linux.efi` → `arch-linux-fallback.efi` before the update
- Kernel PostTransaction hook generates new `arch-linux.efi`
- Result: always have 2 UKIs (current + previous) after first kernel update

### Secure Boot (opt-in via `SecureBoot` config param, default false)
- disabled: zero footprint — sbctl not installed, no keys; sbctl's pacman/mkinitcpio
  hooks ship inside the package, so they are absent too
- enabled: Phase 1 `secureBootSign()` pacstraps sbctl into the target itself
  (all SB setup behind one gate) and signs bootloader copies + UKI; Phase 2
  `secureBoot()` ensures the package, creates keys when missing, keeps all EFI
  binaries signed (`sign -s`, idempotent), hard-errors with fix hints on signing
  failure, prints the activation checklist
- re-signing automation comes with sbctl itself: mkinitcpio post hook
  (`/usr/lib/initcpio/post/sbctl`) signs every rebuild incl. manual mkinitcpio runs;
  pacman transactions additionally get `zz-sbctl.hook`
- **UKI regeneration must always go through mkinitcpio, never raw `ukify`** — raw
  ukify bypasses the post hook and produces unsigned UKIs
- activation/enrollment is manual (firmware ritual); README §Secure Boot is the
  canonical procedure: `sbctl verify` all-green before `enroll-keys -m`


## Network Architecture

### Stack

```
iwd (WiFi daemon) ─── systemd-networkd ─── systemd-resolved
```

- **iwd** — lightweight WiFi daemon, replaces `wpa_supplicant`. Manages WiFi profiles in `/var/lib/iwd/`. Interface naming: keeps kernel name (`wlan0`) via its built-in `80-iwd.link` — no renaming race.
- **systemd-networkd** — handles IP configuration (DHCP) for all interfaces. WiFi matched by `Type=wlan` (agnostic to interface name), ethernet matched by `Name=enp*` (predictable naming).
- **systemd-resolved** — DNS resolution with systemd-resolved configs.

### Phase 1 (ArchISO)

- Uses `iwctl` directly — ArchISO ships iwd pre-installed and running (no separate service management needed).
- `wifiConnect()` in `base.go`: `iwctl --passphrase <password> station <dev> connect <ssid>` (SSID passed as a single argv element via `steps.RunCmd`/`exec.Command` — no shell, so spaces inside an SSID need no quoting; iwctl rejects lowercase-quoted names as invalid network name chars)

### Phase 2 (installed system)

- `network()` in `full.go` writes a `.network` file:
  - **WiFi**: `10-wlan.network` with `[Match] Type=wlan`, `DHCP=yes`, `IgnoreCarrierLoss=3s`.
  - **Ethernet**: `0-eth-dhcp.network` (fixed filename) with `[Match] Name={cfg.NetDev}`.
- `wifiSetup()` in `full.go` drives the connection through `iwd` itself via
  `iwctl --passphrase ... station <dev> connect <ssid>` (same mechanism as Phase 1
  `wifiConnect()`), so **iwd generates and owns the profile** in `/var/lib/iwd/`
  (PreSharedKey/SAE state, AutoConnect) by actually authenticating - never a
  hand-written `.psk`. Also restores `/etc/iwd/main.conf`, enables iwd and
  **starts, never restarts** it (a running daemon keeps the live link; a restart
  is what once dropped the link AND blocked iwd's own restart for ~10min when the
  netdev briefly vanished - the drop-in's device-unit `After=` waits 90s on a
  missing device and the ExecStartPre `ip link set wlan0 up` fails).
- Connect policy is **fail-fast and deterministic**: `wifiSetup()` always re-connects
  to the configured SSID (even when a network is already live - a ~1-3s blip,
  acceptable for a rare command; `zerno i` is re-run to sync, not on a timer).
  A failed connect is a **hard install error with the immediate recovery command
  printed** - bad SSID or password surfaces at sync time instead of poisoning
  autoconnect and dying at the worst possible offline moment later. No state
  parsing of command output (brittle); the behavior is the same every run.
- A stale profile for the configured SSID is forgotten
  (`iwctl known-networks <ssid> forget`) before connecting, so a previously
  unconnectable network can't keep iwd retrying it on every boot.
- `iwctl` is the scriptable CLI used by zerno; `impala` (also in base packages)
  is the interactive TUI for humans. Both are thin D-Bus clients of the same
  iwd daemon, so connecting from either produces the same profile.

### QEMU bridge

- The `qemu0-uplink.network` template uses `[Match] Type=wlan` for WiFi interfaces, `[Match] Name={NetDev}` for ethernet.
- The bridge interface (`qemu0`) gets a static IP + DHCPServer for VMs.

### Key files

| File | Purpose |
|------|---------|
| `internal/install/base.go` | Phase 1 `wifiConnect()` + `pacstrap()` |
| `internal/install/full.go` | Phase 2 `network()` + `wifiSetup()` |
| `internal/install/extra.go` | QEMU uplink |
| `assets/files/iwd-main.conf` | iwd daemon config (`EnableNetworkConfiguration=false`, `NameResolvingService=systemd`) |
| `assets/qemu/uplink.network` | QEMU bridge uplink template |

## Sway/Waybar Nav Widget

Headerless tabs replacement: window icons of the focused workspace rendered in waybar.

- `assets/conf/waybar-nav.py` — resident python daemon (stdlib only), one instance
  per output (`$WAYBAR_OUTPUT_NAME` set by waybar). Subscribes to sway IPC directly
  (framed JSON over `$SWAYSOCK`), renders pango icon spans on window/workspace events.
  No polling; ~1.6 ms per event. Self-culls same-output duplicates via `/proc` scan;
  reconnect loop survives sway restarts. Icon glyphs are PUA codepoints written as
  escapes (literals get stripped in transit) - see `~/src/waybar-icons.md`.
- `assets/conf/waybar.sh` — lifecycle entry, `exec_always` from sway config:
  kills waybar + all `sway/waybar-nav[.]py`, then execs waybar (which spawns one nav per bar).
- waybar module `custom/nav`: `return-type json`, consumes stdout lines; click/scroll
  bindings focus prev/next sibling, middle-click kills.

## Colors Map

No theme engine - colors are literal hex values. To retheme, touch these places
(keep values in sync manually):

| Role | Locations |
|------|-----------|
| bg/output | sway `config`: `$bg #232323`, `$black #000000`; bar bg + tooltip in `waybar.css` |
| gray/element | sway `$gray #3a3a3a`; same hex in `waybar.css` (focused ws, borders, muted) and `waybar.json` separator span; `waybar-nav.py` `CHIP_BG` |
| fg/active | `#ffffff`: sway `$white`; nav `ACTIVE`; css text colors |
| inactive/dim | sway `$dark #202020`; css `#101010` unfocused child_border equivalents live in sway config only |
| accent | sway `$urgent #c25c02` (+ client.urgent row); reused for warnings elsewhere |

Independent palettes NOT covered by the above (own schemes): `alacritty.toml`,
ghostty config, nvim colorscheme (see vim.md).

## Audio (soundSetup / WirePlumber)

- `soundSetup()` in `full.go` installs pipewire/wireplumber/rtkit. rtkit gives audio
  threads realtime priority (SCHED_FIFO) via D-Bus - the standard ArchWiki path; do
  not use the `realtime` group instead. rtkit is D-Bus-activated: it starts on-demand
  when PipeWire first requests realtime scheduling and manages its own lifetime, so
  do NOT `systemctl enable/start rtkit-daemon.service` (leave it disabled, `inactive`
  is the healthy resting state). Verify with
  `busctl --system get-property org.freedesktop.RealtimeKit1 /org/freedesktop/RealtimeKit1 org.freedesktop.RealtimeKit1 MaxRealtimePriority`
  (prints `i 20`; triggers activation if idle) or `ps -eLo comm,cls,rtprio | grep pipewire`
  (look for `FF`).
- Known minor annoyance (not yet acted on): WirePlumber suspends idle audio nodes
  after `session.suspend-timeout-seconds` (default **5s**). On resume, sink/source
  re-activation can chop the first sound or clip the input head - worse over BT.
  Every mainstream distro ships the 5s default; Bazzite sets **0** (never suspend)
  and SteamOS uses 3600 for amp warm-up only. If it ever becomes a real problem,
  drop `/etc/wireplumber/wireplumber.conf.d/51-disable-suspension.conf` with
  `monitor.{alsa,bluez}.properties = { session.suspend-timeout-seconds = 0 }` and
  restart pipewire/wireplumber. `= 0` is the spec'd "never suspend" value, not a
  smaller number. Caveat: `= 0` is necessary but not always sufficient for BT
  (Nixpkgs #528030 residual ~1s delay).

## Design Decisions

- **Reliability over convenience** - prefer predictable behavior that fails
  loudly at sync time over clever/silent shortcuts that delay failure to the
  worst moment (e.g. wifi always reconnects to the configured SSID instead of
  guessing whether to "skip because it's probably fine").
- **No external dependencies** - use stdlib where possible
- **No Makefile** - use `build.sh` instead
- **Binary in repo root** - `zerno`
- **Assets embedded** - configs/templates embedded in binary
- **Repo optional** - only needed for `build-iso`
- **systemd-boot over GRUB** — simpler, modern (UKI, auto-discovery, Secure Boot native), no scripting language
- **MBR/BIOS dropped** — systemd-boot requires UEFI; legacy BIOS is increasingly rare for new installs
- **Single `linux.preset`** — always describes the "active" kernel
- **Always sign Secure Boot keys** — zero user friction, works whether SB is on or off
