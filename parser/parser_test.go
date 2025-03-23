package parser_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheEdgeOfRage/logfmt/config"
	"github.com/TheEdgeOfRage/logfmt/parser"
)

func TestParseLevels(t *testing.T) {
	data := strings.NewReader(`time="2025-03-15T10:32:23Z" level=debug msg="bar"
time="2025-03-15T10:32:24Z" level=info msg="foo"
time="2025-03-15T10:32:25Z" level=warn msg="oopsie"
time="2025-03-15T10:32:26Z" level=error msg="oh no"
time="2025-03-15T10:32:27Z" level=fatal msg="AAAAAA"`)
	w := &bytes.Buffer{}

	p := parser.NewParser(&config.Config{}, data, w)
	err := p.Start()
	require.NoError(t, err)

	assert.Equal(t, `2025-03-15 10:32:23 [DEBUG] msg=bar
2025-03-15 10:32:24  [INFO] msg=foo
2025-03-15 10:32:25  [WARN] msg=oopsie
2025-03-15 10:32:26 [ERROR] msg="oh no"
2025-03-15 10:32:27 [FATAL] msg=AAAAAA
`, w.String())
}

func TestParseTimestamps(t *testing.T) {
	data := strings.NewReader(`timestamp="2025-03-15T10:32:23Z" level=info
time="2025-03-15T10:32:24Z" level=info
ts="2025-03-15T10:32:25Z" level=info
datetime="2025-03-15T10:32:26Z" level=info`)
	w := &bytes.Buffer{}

	p := parser.NewParser(&config.Config{}, data, w)
	err := p.Start()
	require.NoError(t, err)

	assert.Equal(t, `2025-03-15 10:32:23  [INFO]
2025-03-15 10:32:24  [INFO]
2025-03-15 10:32:25  [INFO]
2025-03-15 10:32:26  [INFO]
`, w.String())
}

func TestParseInvalidLogs(t *testing.T) {
	data := strings.NewReader(`
time="2025-03-15T10:32:23Z" level=info msg="loading"
time="`)
	w := &bytes.Buffer{}
	p := parser.NewParser(&config.Config{}, data, w)
	err := p.Start()
	require.Error(t, err)
}
