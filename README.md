# hyprland-share-picker-gtk
A simple replacement of hyprland-share-picker in GTK

## Disclaimer
Warning, this is still a work in progress.
You MUST make a backup of the original binary

## Installation
```bash
go build .
sudo mv /usr/bin/hyprland-share-picker ./hyprland-share-picker-bkup
sudo mv ./hyprland-share-picker-gtk /usr/bin/hyprland-share-picker
```