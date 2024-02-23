package main

import (
	"fmt"
	"os"

	"github.com/TheEdgeOfRage/logfmt/config"
	"github.com/TheEdgeOfRage/logfmt/parser"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	p := parser.NewParser(cfg, os.Stdin, os.Stdout)
	err = p.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
