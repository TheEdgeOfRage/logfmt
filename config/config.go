package config

import (
	"fmt"
	"strings"

	"github.com/go-logfmt/logfmt"
	flags "github.com/jessevdk/go-flags"
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
	// Color is a flag to enable or disable color on the output
	Color bool
}

type rawConfig struct {
	LogLevel     string `long:"level" short:"l" description:"Log level filter. One of DEBUG, INFO, WARN, ERROR, FATAL" default:"INFO"` // nolint:lll
	OutputFields string `long:"output" short:"o" description:"Output field selector (space separated)"`
	Filter       string `long:"filter" short:"f" description:"Filter fields (key=value space separated)"`
	NoColor      bool   `long:"no-color" short:"n" description:"Disable color output"`
}

func Parse() (*Config, error) {
	var raw rawConfig

	_, err := flags.Parse(&raw)
	if err != nil {
		return nil, err
	}

	cfg := Config{
		Color: !raw.NoColor,
	}
	cfg.setOutputFields(raw.OutputFields)
	err = cfg.setFilter(raw.Filter)
	if err != nil {
		return nil, err
	}
	err = cfg.setLevel(raw.LogLevel)
	if err != nil {
		return nil, err
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
	c.OutputFields = strings.Split(fields, " ")
	if len(c.OutputFields) == 1 && c.OutputFields[0] == "" {
		c.OutputFields = nil
	}
}

func (c *Config) setFilter(filter string) error {
	dec := logfmt.NewDecoder(strings.NewReader(filter))
	c.Filter = make(map[string]string)
	dec.ScanRecord()
	for dec.ScanKeyval() {
		if dec.Err() != nil {
			return dec.Err()
		}
		c.Filter[string(dec.Key())] = string(dec.Value())
	}
	return nil
}
