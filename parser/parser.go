package parser

import (
	"bufio"
	"fmt"
	"io"

	"github.com/TheEdgeOfRage/logfmt/config"
)

type Parser struct {
	cfg        *config.Config
	input      io.Reader
	output     io.Writer
	bufferSize int
}

func NewParser(cfg *config.Config, input io.Reader, output io.Writer) *Parser {
	bufferSize := cfg.BufferSize
	if bufferSize == 0 {
		bufferSize = config.DefaultBufferSize
	}
	return &Parser{
		cfg:        cfg,
		input:      input,
		output:     output,
		bufferSize: bufferSize,
	}
}

// Start starts the parser, reading from the input stream and printing the log output to the output stream line by line
func (p *Parser) Start() error {
	scanner := bufio.NewScanner(p.input)
	scanner.Buffer(make([]byte, 0, p.bufferSize), p.bufferSize)
	for scanner.Scan() {
		line := scanner.Text()
		record, err := NewRecord(line, p.cfg, p.bufferSize)
		if err != nil {
			if _, err := fmt.Fprintln(p.output, line); err != nil {
				return fmt.Errorf("failed to print log to output: %w", err)
			}
			continue
		}
		if record.level < p.cfg.LogLevel {
			continue
		}
		if len(p.cfg.Filter) > 0 && !record.MatchesFilter(p.cfg.Filter) {
			continue
		}
		out := record.String(p.cfg)
		if out != "" || p.cfg.KeepEmpty {
			if _, err := fmt.Fprintln(p.output, out); err != nil {
				return fmt.Errorf("failed to print log to output: %w", err)
			}
		}
	}
	return scanner.Err()
}
