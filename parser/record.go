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
	levelStrings    map[int]string
	levelColors     map[int]*color.Color
	timestampLabels = []string{"time", "timestamp", "datetime", "ts", "t"}
)

func init() {
	levelColors = map[int]*color.Color{
		config.Trace:   color.New(color.FgBlack).Add(color.BgHiBlue).Add(color.Bold),
		config.Debug:   color.New(color.FgBlack).Add(color.BgCyan).Add(color.Bold),
		config.Info:    color.New(color.FgBlack).Add(color.BgGreen).Add(color.Bold),
		config.Warning: color.New(color.FgBlack).Add(color.BgYellow).Add(color.Bold),
		config.Error:   color.New(color.FgBlack).Add(color.BgRed).Add(color.Bold),
		config.Fatal:   color.New(color.FgRed).Add(color.BgBlack).Add(color.Bold),
	}
	levelStrings = map[int]string{
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
	level      int
	time       time.Time
	fields     map[string]string
	fieldOrder []string
}

// NewRecord parses a new Record from a logfmt.Decoder
func NewRecord(decoder *logfmt.Decoder, cfg *config.Config) (*Record, error) {
	var record Record
	record.fields = make(map[string]string)
	record.fieldOrder = make([]string, 0)

	for decoder.ScanKeyval() {
		err := decoder.Err()
		if err != nil {
			return nil, fmt.Errorf("failed to parse log line: %w", decoder.Err())
		}
		key, value := string(decoder.Key()), string(decoder.Value())
		if key == "level" {
			record.parseLevel(value)
			if !cfg.Raw {
				continue
			}
		}
		if slices.Contains(timestampLabels, key) {
			err := record.parseTime(value)
			if err != nil {
				return nil, err
			}
			if !cfg.Raw {
				continue
			}
		}
		record.fields[key] = value
		record.fieldOrder = append(record.fieldOrder, key)
	}

	return &record, nil
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
	return s == "null" || s == "NULL" || s == "nil" || s == "<nil>" || s == "None"
}

func getFormattedValue(value string) string {
	if isNumeric(value) || isBoolean(value) {
		return color.MagentaString(value)
	}
	if isNull(value) {
		return color.YellowString(value)
	}
	if strings.Contains(value, " ") {
		value = fmt.Sprintf(`"%s"`, value)
	}
	return color.HiGreenString(value)
}

func getFormattedLevel(level int) string {
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

func (r *Record) parseTime(timeStr string) error {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}

	r.time = t
	return nil
}

func (r *Record) MatchesFilter(filter map[string]string) bool {
	for key, value := range filter {
		if r.fields[key] != value {
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
			reorderedFields := cfg.OutputFields

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

	if color.NoColor {
		fmtString.WriteString("%7s%s")
	} else {
		fmtString.WriteString("%26s%s")
	}

	if cfg.NoTime {
		return fmt.Sprintf(fmtString.String(), getFormattedLevel(r.level), line)
	} else {
		return fmt.Sprintf(fmtString.String(), r.time.Format("2006-01-02 15:04:05"), getFormattedLevel(r.level), line)
	}
}
