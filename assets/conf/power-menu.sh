#!/bin/bash
choice=$(printf "suspend\nhibernate\nlock\nreboot\npoweroff\nexit" | bemenu -b -l 6 -p '▪' --fn 'JetBrainsMono Nerd Font Mono 12' --tb '#c25c02' --tf '#ffffff' --hb '#c25c02' --hf '#ffffff')
case "$choice" in
    suspend)   exec systemctl suspend ;;
    hibernate) exec systemctl hibernate ;;
    lock)      exec swaylock -l -k --color '#000000' ;;
    reboot)    exec systemctl reboot -i ;;
    poweroff)  exec systemctl poweroff -i ;;
    exit)      exec swaymsg exit ;;
esac
