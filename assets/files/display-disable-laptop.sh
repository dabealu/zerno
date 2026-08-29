#!/bin/bash
set -e
EXT_DP=$( swaymsg --pretty -t get_outputs | awk '/Output DP/ {print $2}' | head -1 )
swaymsg -t command output $EXT_DP res 2560x1440@60Hz pos 0 0
swaymsg -t command output $EXT_DP scale 1.0
swaymsg -t command output eDP-1 disable
