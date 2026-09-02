#!/bin/sh

# cap the CPU+iGPU package power draw (Intel RAPL PL1/PL2) on a Framework 13, Iris Xe 96EU.

# WHY:
#   * Default envelope PL1=28W / PL2=64W with the CPU pinned to max boost cooks the package:
#     Intel clamps CPU to 400MHz at ~96-97C (emergency throttle, ~2 min hysteresis) and fps
#     dies mid-game.
#   * 17W: avg 77-79C / max 90-93, mostly 60fps  -> fine but warm.
#   * 13W: avg 68C / max 85C, ~90% of the session in a 65-70C belt and around 50-60fps.
#   * The package sits pinned at the cap the whole time gaming (fully saturated), so the
#     cap is effectively the fps knob.
#   * Undervolting would buy ~5-10C but this chip is fused OC-locked (intel-undervolt writes
#     are ignored: "Values do not equal") - power capping is the only real lever.
#   * Battery: the old 28/64W envelope roasting the chassis; 13W keeps the battery bay cool.
#
# HOW:
#   RAPL exposes PL1 (constraint_0 = sustained) and PL2 (constraint_1 = short-term burst) as
#   writable sysfs files. Both are set to the SAME value here so no turbo burst can exceed the
#   ceiling (bursts are what tripped the cooler at 20W). Takes effect immediately, live - no
#   reboot, no game restart. Resets at next boot unless --persist installs a systemd-tmpfiles
#   rule (writes re-applied every boot automatically).
#   Tip: the CPU governor matters just as much as the watt cap (a 'performance' governor pins
#   every core to max boost and steals the shared package budget from the iGPU).
#
# USAGE (run as root):
#   laptop-tdp-adj.sh 13                  apply 13W to PL1+PL2 (default 13)
#   laptop-tdp-adj.sh 15 --persist        same, at 15W, AND make it survive reboots
#   laptop-tdp-adj.sh --reset             restore the stock 28/64W envelope
#   laptop-tdp-adj.sh --persist           persist the current on-disk tmpfiles value
#   laptop-tdp-adj.sh --show              print current PL1/PL2 + boot persistence (no root)
#   The CPU governor is deliberately NOT touched here. For a quick session flip:
#       cpupower frequency-set -g <governor>
#
# For games on this GPU:
#       gamescope -f -W 2560 -H 1440 -w 1600 -h 900 -F fsr -- %command%
#   Keep the upscaled (-W/-H) pair to your monitor, pick the render target (-w/-h) to taste,
#   add -F fsr on Iris Xe. -r 60 caps frame gen and saves heat. Lighter games might not need
#   the 13W compromise at all.

set -eu
rapl=/sys/class/powercap/intel-rapl:0
# systemd-tmpfiles drop-in: /etc/tmpfiles.d/*.conf is read at every boot by
# systemd-tmpfiles-setup.service (systemd-tmpfiles --create). Each line is
# "TYPE PATH MODE OWNER GROUP AGE ARG"; type 'w' writes ARG into PATH once at
# boot, so sysfs knobs that don't survive reboot get re-applied automatically.
# /etc shadows /run which shadows /usr/lib for same-named files. This file is
# also honored on-demand via `systemd-tmpfiles --create`.
tmpfiles=/etc/tmpfiles.d/rapl-power-limits.conf
default_w=13
persist=0
reset=0
show=0

for arg in "$@"; do
    case "$arg" in
        --persist) persist=1 ;;
        --reset)   reset=1 ;;
        --show)    show=1 ;;
        *)
            case "$arg" in
                *[!0-9]*) echo "ignoring bad arg: '$arg'" >&2 ;;
                *) w="$arg" ;;
            esac
            ;;
    esac
done

# Show-only needs no privileges beyond reading sysfs; bail before the root check.
if [ "$show" -eq 1 ]; then
    if [ ! -r "$rapl/constraint_0_power_limit_uw" ]; then
        echo "no RAPL PL1 knob at $rapl - run this only on the laptop it was written for"
        exit 1
    fi
    echo "current: PL1=$(cat "$rapl/constraint_0_power_limit_uw")uW / PL2=$(cat "$rapl/constraint_1_power_limit_uw")uW"
    if [ -f "$tmpfiles" ]; then
        echo "boot persistence: yes ($tmpfiles)"
        echo "  will apply at boot: $(cat "$tmpfiles")"
    else
        echo "boot persistence: no (resets to firmware defaults on reboot)"
    fi
    exit 0
fi

if [ "$(id -u)" -ne 0 ]; then
    echo "run as root (sudo $0 $*)"
    exit 1
fi

if [ ! -r "$rapl/constraint_0_power_limit_uw" ]; then
    echo "no RAPL PL1 knob at $rapl - run this only on the laptop it was written for"
    exit 1
fi

if [ "$reset" -eq 1 ]; then
    pl1=28000000; pl2=64000000; tag="stock 28W / 64W"
    rm -f "$tmpfiles"
    echo "removed boot persistence ($tmpfiles)"
else
    w="${w:-$default_w}"
    pl1=$(($w * 1000000)); pl2=$pl1; tag="$w W sustained"
fi

echo "$pl1" >"$rapl/constraint_0_power_limit_uw"
echo "$pl2" >"$rapl/constraint_1_power_limit_uw"

if [ "$persist" -eq 1 ] && [ "$reset" -ne 1 ]; then
    printf 'w %s/constraint_0_power_limit_uw - - - - %s\n' "$rapl" "$pl1" >"$tmpfiles"
    printf 'w %s/constraint_1_power_limit_uw - - - - %s\n' "$rapl" "$pl2" >>"$tmpfiles"
    echo "persisted: $tmpfiles = $w W"
fi

echo "applied: $tag (PL1=$(cat "$rapl/constraint_0_power_limit_uw") uW PL2=$(cat "$rapl/constraint_1_power_limit_uw") uW)"
echo
echo "bake into boot:  $0 $w --persist"

