package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/go-logfmt/logfmt"
)

const (
	Debug   = "DEBUG"
	Info    = "INFO"
	Warning = "WARN"
	Error   = "ERROR"
	Fatal   = "FATAL"
)

var levelColors map[string]*color.Color

func init() {
	levelColors = map[string]*color.Color{
		Debug:   color.New(color.FgBlack).Add(color.BgCyan).Add(color.Bold),
		Info:    color.New(color.FgBlack).Add(color.BgGreen).Add(color.Bold),
		Warning: color.New(color.FgBlack).Add(color.BgYellow).Add(color.Bold),
		Error:   color.New(color.FgBlack).Add(color.BgRed).Add(color.Bold),
		Fatal:   color.New(color.FgBlack).Add(color.BgRed).Add(color.Bold),
	}
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

func getLevel(level string) string {
	level = strings.ToUpper(level)
	if level == "WARNING" {
		level = Warning
	}
	val, ok := levelColors[level]
	if !ok {
		return levelColors[Info].Sprintf("[%s]", Info)
	}
	return val.Sprintf("[%s]", level)
}

func getValue(value string) string {
	if isNumeric(value) || isBoolean(value) {
		return color.MagentaString(value)
	}
	if isNull(value) {
		return color.YellowString(value)
	}
	return color.HiGreenString(value)
}

func main() {
	decoder := logfmt.NewDecoder(os.Stdin)
	for decoder.ScanRecord() {
		line := ""
		level := Info
		timestamp := ""
		for decoder.ScanKeyval() {
			key, value := string(decoder.Key()), string(decoder.Value())
			if key == "level" {
				level = getLevel(value)
				continue
			}
			if key == "time" || key == "timestamp" {
				timestamp = value
				continue
			}
			line += fmt.Sprintf("%s=%s ", color.HiBlueString(key), getValue(value))
		}

		fmt.Printf("%s %26s %s\n", timestamp, level, line)
	}
}
