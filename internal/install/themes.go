package install

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"zerno/assets"
	"zerno/internal/paths"
)

type themeFile struct {
	path  string
	asset string
}

var hexRe = regexp.MustCompile(`#[0-9a-fA-F]{6}`)

func defaultPaths(homeDir string) []themeFile {
	swayDir := filepath.Join(homeDir, ".config/sway")
	return []themeFile{
		{path: filepath.Join(homeDir, ".config/dunst/dunstrc"), asset: "conf/dunstrc"},
		{path: filepath.Join(swayDir, "config"), asset: "conf/config"},
		{path: filepath.Join(swayDir, "waybar.css"), asset: "conf/waybar.css"},
		{path: filepath.Join(swayDir, "power-menu.sh"), asset: "conf/power-menu.sh"},
		{path: filepath.Join(swayDir, "fav-apps.sh"), asset: "conf/fav-apps.sh"},
	}
}

func restoreConfigs() error {
	for _, f := range defaultPaths(paths.HomeDir()) {
		log.Printf("restoring %s -> %s", f.asset, f.path)
		if err := assets.Restore(f.asset, f.path); err != nil {
			return err
		}
	}
	return nil
}

func ApplyTheme(palette map[string]string) error {
	if err := restoreConfigs(); err != nil {
		return err
	}

	for _, f := range defaultPaths(paths.HomeDir()) {
		data, err := os.ReadFile(f.path)
		if err != nil {
			return err
		}
		info, err := os.Stat(f.path)
		if err != nil {
			return err
		}
		s := hexRe.ReplaceAllStringFunc(string(data), func(match string) string {
			if theme, ok := palette[match]; ok {
				return theme
			}
			return match
		})
		if err := os.WriteFile(f.path, []byte(s), info.Mode()); err != nil {
			return err
		}
	}

	return reloadServices()
}

func ThemeDefault() error {
	if err := restoreConfigs(); err != nil {
		return err
	}
	return reloadServices()
}

func reloadServices() error {
	exec.Command("pkill", "waybar").Run()
	exec.Command("pkill", "dunst").Run()
	exec.Command("swaymsg", "reload").Run()
	return nil
}

// Themes is a map of available palettes
var Themes = map[string]map[string]string{
	"default": nil, // added to show it in available themes list
	"suede": { // blue-warm suede
		"#000000": "#151d2b", // bg
		"#ffffff": "#E7C096", // fg
		"#404040": "#25375F", // element
		"#202020": "#1a2638", // inactive
		"#232323": "#1e2a40", // output
		"#c25c02": "#3C5899", // accent
		"#777777": "#C89468", // title
		"#808080": "#A88260", // secondary
		"#101010": "#1a2638", // dunst-bg
		"#301010": "#2a1a28", // dunst-crit
		"#999999": "#2F4883", // dunst-frame
		"#ffead3": "#E7C096", // time
		"#99ffdd": "#3C5899", // green
		"#ffcc66": "#CE9C70", // yellow
		"#00ff00": "#4A65A2", // green-alt
		"#00FF00": "#E7C096", // cursor
	},
	"gruvbox-dark": { // warm retro
		"#000000": "#282828", // bg
		"#ffffff": "#ebdbb2", // fg
		"#404040": "#504945", // element
		"#202020": "#3c3836", // inactive
		"#232323": "#45403d", // output
		"#c25c02": "#d65d0e", // accent
		"#777777": "#665c54", // title
		"#808080": "#7c6f64", // secondary
		"#101010": "#3c3836", // dunst-bg
		"#301010": "#4a2e2a", // dunst-crit
		"#999999": "#504945", // dunst-frame
		"#ffead3": "#ebdbb2", // time
		"#99ffdd": "#689d6a", // green
		"#ffcc66": "#d79921", // yellow
		"#00ff00": "#98971a", // green-alt
		"#00FF00": "#ebdbb2", // cursor
	},
	"wilderness": { // earthy brown-green
		"#000000": "#1b1918", // bg
		"#ffffff": "#dbd1b8", // fg
		"#404040": "#3e3934", // element (lighter for contrast)
		"#202020": "#262320", // inactive
		"#232323": "#2a2623", // output
		"#c25c02": "#5c6e5f", // accent (darker for bemenu)
		"#777777": "#6b635c", // title
		"#808080": "#7d736a", // secondary
		"#101010": "#262320", // dunst-bg
		"#301010": "#3a2826", // dunst-crit
		"#999999": "#2f2b28", // dunst-frame
		"#ffead3": "#dbd1b8", // time
		"#99ffdd": "#6b9b8a", // green
		"#ffcc66": "#b89a6b", // yellow
		"#00ff00": "#768b7a", // green-alt
		"#00FF00": "#dbd1b8", // cursor
	},
	"ayu-dark": { // deep indigo with warm gold accent
		"#000000": "#0d1017", // bg
		"#ffffff": "#d4d4d4", // fg
		"#404040": "#2d3f52", // element
		"#202020": "#11141c", // inactive
		"#232323": "#151a23", // output
		"#c25c02": "#a07a28", // accent
		"#777777": "#565b66", // title
		"#808080": "#6c7180", // secondary
		"#101010": "#11141c", // dunst-bg
		"#301010": "#261a1a", // dunst-crit
		"#999999": "#1a1f29", // dunst-frame
		"#ffead3": "#d4d4d4", // time
		"#99ffdd": "#4dbf99", // green
		"#ffcc66": "#a07a28", // yellow
		"#00ff00": "#aad94c", // green-alt
		"#00FF00": "#d4d4d4", // cursor
	},
	"ember": { // warm ember dark (ember-theme)
		"#000000": "#1c1b19", // bg
		"#ffffff": "#d8d0c0", // fg
		"#404040": "#3e3c38", // element
		"#202020": "#242320", // inactive
		"#232323": "#2e2d2a", // output
		"#c25c02": "#e08060", // accent (coral)
		"#777777": "#b8b0a0", // title
		"#808080": "#706c61", // secondary
		"#101010": "#1c1b19", // dunst-bg
		"#301010": "#6a2e1e", // dunst-crit (dark coral)
		"#999999": "#2e2d2a", // dunst-frame
		"#ffead3": "#d8d0c0", // time
		"#99ffdd": "#80a090", // green (sage)
		"#ffcc66": "#c8b468", // yellow (gold)
		"#00ff00": "#8a9868", // green-alt (olive)
		"#00FF00": "#d8d0c0", // cursor
	},
}
