package parser

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/TheEdgeOfRage/logfmt/config"
	"github.com/fatih/color"
	"github.com/go-logfmt/logfmt"
)

var (
	levelStrings map[int]string
	levelColors  map[int]*color.Color
)

func init() {
	levelColors = map[int]*color.Color{
		config.Debug:   color.New(color.FgBlack).Add(color.BgCyan).Add(color.Bold),
		config.Info:    color.New(color.FgBlack).Add(color.BgGreen).Add(color.Bold),
		config.Warning: color.New(color.FgBlack).Add(color.BgYellow).Add(color.Bold),
		config.Error:   color.New(color.FgBlack).Add(color.BgRed).Add(color.Bold),
		config.Fatal:   color.New(color.FgRed).Add(color.BgBlack).Add(color.Bold),
	}
	levelStrings = map[int]string{
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
func NewRecord(decoder *logfmt.Decoder) (*Record, error) {
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
			continue
		}
		if key == "time" || key == "timestamp" {
			err := record.parseTime(value)
			if err != nil {
				return nil, err
			}
			continue
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
	case "DEBUG":
		r.level = config.Debug
	case "INFO":
		r.level = config.Info
	case "WARN", "WARNING":
		r.level = config.Warning
	case "ERROR":
		r.level = config.Error
	case "FATAL":
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
	for _, key := range r.fieldOrder {
		if len(cfg.OutputFields) > 0 && !slices.Contains(cfg.OutputFields, key) {
			continue
		}
		if len(cfg.ExcludeFields) > 0 && slices.Contains(cfg.ExcludeFields, key) {
			continue
		}
		value := r.fields[key]
		key = color.HiBlueString(key)
		line += fmt.Sprintf("%s=%s ", key, getFormattedValue(value))
	}

	var fmtString strings.Builder
	if !cfg.NoTime {
		fmtString.WriteString("%s ")
	}

	if color.NoColor {
		fmtString.WriteString("%7s %s")
	} else {
		fmtString.WriteString("%26s %s")
	}

	if cfg.NoTime {
		return fmt.Sprintf(fmtString.String(), getFormattedLevel(r.level), line)
	} else {
		return fmt.Sprintf(fmtString.String(), r.time.Format("2006-01-02 15:04:05"), getFormattedLevel(r.level), line)
	}
}
