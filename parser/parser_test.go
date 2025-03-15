package parser_test

import (
	"os"
	"testing"

	"github.com/TheEdgeOfRage/logfmt/config"
	"github.com/TheEdgeOfRage/logfmt/parser"
)

func TestParseInvalidFile(t *testing.T) {
	f, err := os.Open("../testdata/log.txt")
	if err != nil {
		t.Fatal(err)
	}
	p := parser.NewParser(&config.Config{}, f, os.Stdout)
	if err := p.Start(); err == nil {
		t.Error("expected error, got nil")
	}
}
