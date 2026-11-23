package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
)

const (
	configDirName  = "logfmt"
	configFileName = "config.toml"
)

// Style is an optional terminal style. The zero value leaves text unchanged.
type Style struct {
	sgrCodes []string
}

func (s Style) Sprint(text string) string {
	if len(s.sgrCodes) == 0 || color.NoColor {
		return text
	}
	return "\x1b[" + strings.Join(s.sgrCodes, ";") + "m" + text + "\x1b[0m"
}

type Colors struct {
	Equal   Style
	Date    Style
	Key     Style
	Null    Style
	Numeric Style
	String  Style
	Levels  [Fatal + 1]Style
}

type colorFile struct {
	Color struct {
		Equal   string `toml:"equal"`
		Date    string `toml:"date"`
		Key     string `toml:"key"`
		Null    string `toml:"null"`
		Numeric string `toml:"numeric"`
		String  string `toml:"string"`
		Levels  struct {
			Trace   string `toml:"trace"`
			Debug   string `toml:"debug"`
			Info    string `toml:"info"`
			Warning string `toml:"warning"`
			Error   string `toml:"error"`
			Fatal   string `toml:"fatal"`
		} `toml:"levels"`
	} `toml:"color"`
}

func defaultColorFile() colorFile {
	var f colorFile
	f.Color.Equal = "normal"
	f.Color.Date = "normal"
	f.Color.Key = "brightblue"
	f.Color.Null = "yellow"
	f.Color.Numeric = "magenta"
	f.Color.String = "brightgreen"
	f.Color.Levels.Trace = "black brightblue bold"
	f.Color.Levels.Debug = "black cyan bold"
	f.Color.Levels.Info = "black green bold"
	f.Color.Levels.Warning = "black yellow bold"
	f.Color.Levels.Error = "black red bold"
	f.Color.Levels.Fatal = "red black bold"
	return f
}

func configFilePath() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, configDirName, configFileName), nil
}

func loadColors() (Colors, error) {
	file := defaultColorFile()
	path, err := configFilePath()
	if err == nil {
		var meta toml.MetaData
		meta, err = toml.DecodeFile(path, &file)
		if err != nil && !os.IsNotExist(err) {
			return Colors{}, fmt.Errorf("reading %s: %w", path, err)
		}
		if undecoded := meta.Undecoded(); len(undecoded) > 0 {
			return Colors{}, fmt.Errorf("%s: unknown keys: %v", path, undecoded)
		}
	}
	return file.colors()
}

func (f colorFile) colors() (Colors, error) {
	var c Colors
	specs := []struct {
		dest *Style
		spec string
		name string
	}{
		{&c.Equal, f.Color.Equal, "equal"},
		{&c.Date, f.Color.Date, "date"},
		{&c.Key, f.Color.Key, "key"},
		{&c.Null, f.Color.Null, "null"},
		{&c.Numeric, f.Color.Numeric, "numeric"},
		{&c.String, f.Color.String, "string"},
		{&c.Levels[Trace], f.Color.Levels.Trace, "levels.trace"},
		{&c.Levels[Debug], f.Color.Levels.Debug, "levels.debug"},
		{&c.Levels[Info], f.Color.Levels.Info, "levels.info"},
		{&c.Levels[Warning], f.Color.Levels.Warning, "levels.warning"},
		{&c.Levels[Error], f.Color.Levels.Error, "levels.error"},
		{&c.Levels[Fatal], f.Color.Levels.Fatal, "levels.fatal"},
	}
	for _, s := range specs {
		style, err := parseStyle(s.spec)
		if err != nil {
			return Colors{}, fmt.Errorf("color.%s: %w", s.name, err)
		}
		*s.dest = style
	}
	return c, nil
}

var baseColors = map[string]int{
	"black": 0, "red": 1, "green": 2, "yellow": 3,
	"blue": 4, "magenta": 5, "cyan": 6, "white": 7,
}

var textStyles = map[string]int{
	"bold":      1,
	"italic":    3,
	"underline": 4,
}

const (
	foregroundBase       = 30
	brightForegroundBase = 90
	backgroundOffset     = 10
	extendedForeground   = 38
	truecolorMode        = 2
)

// parseStyle parses "foreground [background [style]]".
func parseStyle(spec string) (Style, error) {
	parts := strings.Fields(spec)
	if len(parts) == 0 || len(parts) > 3 {
		return Style{}, fmt.Errorf("invalid style %q, want \"foreground [background [style]]\"", spec)
	}

	var codes []string
	for i, part := range parts[:min(len(parts), 2)] {
		colorAttrs, err := parseColor(part, i == 1)
		if err != nil {
			return Style{}, err
		}
		codes = append(codes, colorAttrs...)
	}
	if len(parts) == 3 && parts[2] != "normal" {
		attr, ok := textStyles[parts[2]]
		if !ok {
			return Style{}, fmt.Errorf("invalid text style %q", parts[2])
		}
		codes = append(codes, strconv.Itoa(attr))
	}
	return Style{sgrCodes: codes}, nil
}

func parseColor(name string, background bool) ([]string, error) {
	offset := 0
	if background {
		offset = backgroundOffset
	}
	if name == "normal" {
		return nil, nil
	}
	if hex, ok := strings.CutPrefix(name, "#"); ok {
		value, err := strconv.ParseUint(hex, 16, 32)
		if err != nil || len(hex) != 6 {
			return nil, fmt.Errorf("invalid hex color %q", name)
		}
		return []string{
			strconv.Itoa(extendedForeground + offset), strconv.Itoa(truecolorMode),
			strconv.Itoa(int(value >> 16)), strconv.Itoa(int(value >> 8 & 0xff)), strconv.Itoa(int(value & 0xff)),
		}, nil
	}
	base := foregroundBase
	if bright, ok := strings.CutPrefix(name, "bright"); ok {
		name = bright
		base = brightForegroundBase
	}
	index, ok := baseColors[name]
	if !ok {
		return nil, fmt.Errorf("invalid color %q", name)
	}
	return []string{strconv.Itoa(base + index + offset)}, nil
}
