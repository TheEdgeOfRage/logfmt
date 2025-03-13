package parser

import (
	"fmt"
	"io"

	"github.com/TheEdgeOfRage/logfmt/config"
	"github.com/go-logfmt/logfmt"
)

type Parser struct {
	cfg    *config.Config
	input  io.Reader
	output io.Writer
}

func NewParser(cfg *config.Config, input io.Reader, output io.Writer) *Parser {
	return &Parser{
		cfg:    cfg,
		input:  input,
		output: output,
	}
}

// Start starts the parser, reading from the input stream and printing the log output to the output stream line by line
func (p *Parser) Start() error {
	decoder := logfmt.NewDecoder(p.input)
	for decoder.ScanRecord() {
		record, err := NewRecord(decoder)
		if err != nil {
			return err
		}
		if record.level < p.cfg.LogLevel {
			continue
		}
		if len(p.cfg.Filter) > 0 && !record.MatchesFilter(p.cfg.Filter) {
			continue
		}

		_, err = fmt.Fprintf(p.output, "%s\n", record.String(p.cfg))
		if err != nil {
			return fmt.Errorf("failed to print log to output: %w", err)
		}
	}

	return nil
}
