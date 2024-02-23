package config

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
)

const (
	Debug = iota
	Info
	Warning
	Error
	Fatal
)

type Config struct {
	// LogLevel is the level filter for the log output
	LogLevel int
	// LogFormat is a list of fields to include on the log output
	OutputFields []string
	// Filter is a map of fields and values which are used to filter the log output
	Filter map[string]string
}

type rawConfig struct {
	LogLevel     string `long:"level" short:"l" description:"Log level filter. One of DEBUG, INFO, WARN, ERROR, FATAL" default:"INFO"` // nolint:lll
	OutputFields string `long:"output" short:"o" description:"Output field selector (comma separated)"`
	Filter       string `long:"filter" short:"f" description:"Filter fields (key=value comma separated)"`
	NoColor      bool   `long:"no-color" short:"n" description:"Disable color output"`
	ForceColor   bool   `long:"force-color" short:"c" description:"Force color output, even when outputting to a pipe"`
}

func Parse() (*Config, error) {
	var raw rawConfig

	parser := flags.NewParser(&raw, flags.HelpFlag|flags.PassDoubleDash)
	_, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	cfg := Config{}
	cfg.setOutputFields(raw.OutputFields)
	err = cfg.setFilter(raw.Filter)
	if err != nil {
		return nil, err
	}
	err = cfg.setLevel(raw.LogLevel)
	if err != nil {
		return nil, err
	}
	if raw.ForceColor && raw.NoColor {
		return nil, fmt.Errorf("cannot use both --force-color and --no-color")
	}
	if raw.ForceColor {
		color.NoColor = false
	}
	if raw.NoColor {
		color.NoColor = true
	}

	return &cfg, nil
}

func (c *Config) setLevel(level string) error {
	switch level {
	case "DEBUG":
		c.LogLevel = Debug
	case "INFO":
		c.LogLevel = Info
	case "WARN":
		c.LogLevel = Warning
	case "ERROR":
		c.LogLevel = Error
	case "FATAL":
		c.LogLevel = Fatal
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
	return nil
}

func (c *Config) setOutputFields(fields string) {
	fields = strings.Trim(fields, " ")
	if fields == "" {
		return
	}
	c.OutputFields = strings.Split(fields, ",")
}

func (c *Config) setFilter(filter string) error {
	if filter == "" {
		return nil
	}

	filters := strings.Split(filter, ",")
	c.Filter = make(map[string]string)
	for _, f := range filters {
		f = strings.Trim(f, " ")
		parts := strings.Split(f, "=")
		if len(parts) != 2 {
			return fmt.Errorf("invalid filter: %s", f)
		}
		c.Filter[parts[0]] = parts[1]
	}
	return nil
}
