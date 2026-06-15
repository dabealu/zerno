#!/bin/bash
choice=$(printf "\
chrome         -  web browser\n\
thunar         -  file manager\n\
telegram       -  messaging\n\
evince         -  pdf documents\n\
ristretto      -  image viewer\n\
transmission   -  torrents\n\
vlc            -  video player\n\
audacious      -  audio player\n\
pavucontrol    -  audio mixer\n\
blueman        -  bluetooth\n\
impala         -  wifi\n\
virt-manager   -  vms\n\
steam          -  games\n\
nwg-look       -  ui appearance\n\
wdisplays      -  display config" | bemenu -b -l 15 -p '★' --fn 'JetBrainsMono Nerd Font Mono 12' --tb '#c25c02' --tf '#ffffff' --hb '#c25c02' --hf '#ffffff')
case "$choice" in
    chrome*)       exec google-chrome-stable ;;
    thunar*)       exec thunar ;;
    telegram*)     exec telegram-desktop ;;
    evince*)       exec evince ;;
    ristretto*)    exec ristretto ;;
    transmission*) exec transmission-gtk ;;
    vlc*)          exec vlc ;;
    audacious*)    exec audacious ;;
    pavucontrol*)  exec pavucontrol ;;
    blueman*)      exec blueman-manager ;;
    impala*)       exec alacritty --config-file ~/.config/sway/alacritty.toml -e impala ;;
    virt-manager*) exec virt-manager ;;
    steam*)        exec steam ;;
    nwg-look*)     exec nwg-look ;;
    wdisplays*)    exec wdisplays ;;
esac
