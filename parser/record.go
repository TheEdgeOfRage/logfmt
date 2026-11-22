package parser

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/go-logfmt/logfmt"

	"github.com/TheEdgeOfRage/logfmt/config"
)

var (
	levelStrings    map[config.Level]string
	levelColors     map[config.Level]*color.Color
	timestampLabels = []string{"time", "timestamp", "datetime", "ts", "t"}
)

func init() {
	levelColors = map[config.Level]*color.Color{
		config.Trace:   color.New(color.FgBlack).Add(color.BgHiBlue).Add(color.Bold),
		config.Debug:   color.New(color.FgBlack).Add(color.BgCyan).Add(color.Bold),
		config.Info:    color.New(color.FgBlack).Add(color.BgGreen).Add(color.Bold),
		config.Warning: color.New(color.FgBlack).Add(color.BgYellow).Add(color.Bold),
		config.Error:   color.New(color.FgBlack).Add(color.BgRed).Add(color.Bold),
		config.Fatal:   color.New(color.FgRed).Add(color.BgBlack).Add(color.Bold),
	}
	levelStrings = map[config.Level]string{
		config.Trace:   "TRACE",
		config.Debug:   "DEBUG",
		config.Info:    "INFO",
		config.Warning: "WARN",
		config.Error:   "ERROR",
		config.Fatal:   "FATAL",
	}
}

// Record represents a single log line
type Record struct {
	level      config.Level
	time       time.Time
	fields     map[string]string
	fieldOrder []string
}

// NewRecord parses a new Record from a single log line
func NewRecord(line string, cfg *config.Config, bufferSize int) (*Record, error) {
	record := &Record{
		level:      config.Info,
		fields:     make(map[string]string),
		fieldOrder: make([]string, 0),
	}

	decoder := logfmt.NewDecoderSize(strings.NewReader(line), bufferSize)
	if !decoder.ScanRecord() {
		return record, nil
	}

	for decoder.ScanKeyval() {
		key, value := string(decoder.Key()), string(decoder.Value())
		if key == "level" {
			record.parseLevel(value)
			if !cfg.Raw {
				continue
			}
		}
		if slices.Contains(timestampLabels, key) {
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				record.time = t
				if !cfg.Raw {
					continue
				}
			}
			// otherwise: treat the field as any other
		}
		record.fields[key] = value
		if !slices.Contains(record.fieldOrder, key) {
			record.fieldOrder = append(record.fieldOrder, key)
		}
	}

	if err := decoder.Err(); err != nil {
		return nil, err
	}

	return record, nil
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isBoolean(s string) bool {
	_, err := strconv.ParseBool(s)
	return err == nil
}

func isNull(s string) bool {
	switch strings.ToLower(s) {
	case "null", "nil", "<nil>", "none":
		return true
	}
	return false
}

func getFormattedValue(value string) string {
	if isNumeric(value) || isBoolean(value) {
		return color.MagentaString(value)
	}
	if isNull(value) {
		return color.YellowString(value)
	}
	if strings.ContainsAny(value, " \"") {
		value = "\"" + strings.ReplaceAll(value, "\"", "\\\"") + "\""
	}
	return color.HiGreenString(value)
}

func getFormattedLevel(level config.Level) string {
	return levelColors[level].Sprintf("[%s]", levelStrings[level])
}

func (r *Record) parseLevel(level string) {
	switch strings.ToUpper(level) {
	case "TRACE":
		r.level = config.Trace
	case "DBUG", "DEBUG":
		r.level = config.Debug
	case "INFO", "INFORMATION", "INFORMATIONAL", "NOTICE":
		r.level = config.Info
	case "WARN", "WARNING":
		r.level = config.Warning
	case "ERR", "EROR", "ERROR":
		r.level = config.Error
	case "EMERG", "FATAL", "ALERT", "CRIT", "CRITICAL":
		r.level = config.Fatal
	default:
		r.level = config.Info
	}
}

func (r *Record) MatchesFilter(filter map[string]string) bool {
	for key, value := range filter {
		actual, ok := r.fields[key]
		if !ok || actual != value {
			return false
		}
	}
	return true
}

// String returns a formatted string representation of the Record
func (r *Record) String(cfg *config.Config) string {
	line := ""

	outFields := r.fieldOrder
	if len(cfg.OutputFields) > 0 {
		if cfg.All {
			reorderedFields := make([]string, 0, len(cfg.OutputFields)+len(r.fieldOrder))
			reorderedFields = append(reorderedFields, cfg.OutputFields...)
			for _, key := range r.fieldOrder {
				if !slices.Contains(reorderedFields, key) {
					reorderedFields = append(reorderedFields, key)
				}
			}
			outFields = reorderedFields
		} else {
			outFields = cfg.OutputFields
		}
	}

	for _, key := range outFields {
		if len(cfg.ExcludeFields) > 0 && slices.Contains(cfg.ExcludeFields, key) {
			continue
		}
		value, ok := r.fields[key]
		if !ok {
			continue
		}
		key := color.HiBlueString(key)
		if cfg.Raw {
			line += fmt.Sprintf(" %s", value)
		} else {
			line += fmt.Sprintf(" %s=%s", key, getFormattedValue(value))
		}
	}

	if line == "" && !cfg.KeepEmpty {
		return ""
	}

	if cfg.Raw {
		return strings.TrimSpace(line)
	}

	var fmtString strings.Builder
	if !cfg.NoTime {
		fmtString.WriteString("%s ")
	}

	if cfg.NoColor {
		fmtString.WriteString("%7s%s")
	} else {
		fmtString.WriteString("%26s%s")
	}

	if cfg.NoTime {
		return fmt.Sprintf(fmtString.String(), getFormattedLevel(r.level), line)
	}

	ts := strings.Repeat(" ", 19)
	if !r.time.IsZero() {
		ts = r.time.Format("2006-01-02 15:04:05")
	}
	return fmt.Sprintf(fmtString.String(), ts, getFormattedLevel(r.level), line)
}
