package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fatih/color"
)

func TestParseStyle(t *testing.T) {
	color.NoColor = false
	tests := map[string]string{
		"normal":                "x",
		"red":                   "\x1b[31mx\x1b[0m",
		"brightblue":            "\x1b[94mx\x1b[0m",
		"black cyan bold":       "\x1b[30;46;1mx\x1b[0m",
		"#ff8000 normal":        "\x1b[38;2;255;128;0mx\x1b[0m",
		"normal #000001 italic": "\x1b[48;2;0;0;1;3mx\x1b[0m",
	}
	for spec, want := range tests {
		style, err := parseStyle(spec)
		if err != nil {
			t.Fatalf("%q: unexpected error: %v", spec, err)
		}
		if got := style.Sprint("x"); got != want {
			t.Errorf("%q: got %q, want %q", spec, got, want)
		}
	}
}

func TestParseStyleInvalid(t *testing.T) {
	for _, spec := range []string{"", "purple", "red blue sparkly", "#12345"} {
		if _, err := parseStyle(spec); err == nil {
			t.Errorf("%q: expected error", spec)
		}
	}
}

func TestLoadColorsFromFile(t *testing.T) {
	color.NoColor = false
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if err := os.MkdirAll(filepath.Join(dir, "logfmt"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "[color]\nkey = \"red\"\n[color.levels]\nerror = \"white\"\n"
	if err := os.WriteFile(filepath.Join(dir, "logfmt", "config.toml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	colors, err := loadColors()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := colors.Key.Sprint("x"); got != "\x1b[31mx\x1b[0m" {
		t.Errorf("key override not applied: %q", got)
	}
	if got := colors.Levels[Error].Sprint("x"); got != "\x1b[37mx\x1b[0m" {
		t.Errorf("error level override not applied: %q", got)
	}
	if got := colors.Null.Sprint("x"); got != "\x1b[33mx\x1b[0m" {
		t.Errorf("default not kept: %q", got)
	}
}
