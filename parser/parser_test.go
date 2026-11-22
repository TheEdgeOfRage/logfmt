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
	data := strings.NewReader(`time="2025-03-15T10:32:22Z" level=trace msg="foobar"
time="2025-03-15T10:32:23Z" level=debug msg="bar"
time="2025-03-15T10:32:24Z" level=info msg="foo"
time="2025-03-15T10:32:25Z" level=warn msg="oopsie"
time="2025-03-15T10:32:26Z" level=error msg="oh no"
time="2025-03-15T10:32:27Z" level=fatal msg="AAAAAA"`)
	w := &bytes.Buffer{}

	p := parser.NewParser(&config.Config{NoColor: true}, data, w)
	err := p.Start()
	require.NoError(t, err)

	assert.Equal(t, `2025-03-15 10:32:22 [TRACE] msg=foobar
2025-03-15 10:32:23 [DEBUG] msg=bar
2025-03-15 10:32:24  [INFO] msg=foo
2025-03-15 10:32:25  [WARN] msg=oopsie
2025-03-15 10:32:26 [ERROR] msg="oh no"
2025-03-15 10:32:27 [FATAL] msg=AAAAAA
`, w.String())
}

func TestParseTimestamps(t *testing.T) {
	data := strings.NewReader(`timestamp="2025-03-15T10:32:23Z" level=info msg=foo
time="2025-03-15T10:32:24Z" level=info msg=foo
ts="2025-03-15T10:32:25Z" level=info msg=foo
datetime="2025-03-15T10:32:26Z" level=info msg=foo`)
	w := &bytes.Buffer{}

	p := parser.NewParser(&config.Config{NoColor: true}, data, w)
	err := p.Start()
	require.NoError(t, err)

	assert.Equal(t, `2025-03-15 10:32:23  [INFO] msg=foo
2025-03-15 10:32:24  [INFO] msg=foo
2025-03-15 10:32:25  [INFO] msg=foo
2025-03-15 10:32:26  [INFO] msg=foo
`, w.String())
}

func TestDuplicateKeysPrintOnce(t *testing.T) {
	data := strings.NewReader(`time="2025-03-15T10:32:23Z" level=info msg=hello id=1 msg=world`)
	w := &bytes.Buffer{}
	p := parser.NewParser(&config.Config{NoColor: true}, data, w)
	err := p.Start()
	require.NoError(t, err)
	assert.Equal(t, "2025-03-15 10:32:23  [INFO] msg=world id=1\n", w.String())
}

func TestBadTimestampDoesNotAbortStream(t *testing.T) {
	data := strings.NewReader(`time="2025-03-15T10:32:23Z" level=info msg=good
time="not-a-timestamp" level=info msg=bad
time="2025-03-15T10:32:25Z" level=info msg=good2`)
	w := &bytes.Buffer{}
	p := parser.NewParser(&config.Config{NoColor: true}, data, w)
	err := p.Start()
	require.NoError(t, err)
	assert.Contains(t, w.String(), "msg=good")
	assert.Contains(t, w.String(), "msg=bad")
	assert.Contains(t, w.String(), "msg=good2")
}

func TestMissingLevelDefaultsToInfo(t *testing.T) {
	data := strings.NewReader(`msg=hello id=1`)
	w := &bytes.Buffer{}
	p := parser.NewParser(&config.Config{NoColor: true}, data, w)
	err := p.Start()
	require.NoError(t, err)
	assert.Contains(t, w.String(), "[INFO]")
	assert.Contains(t, w.String(), "msg=hello")
}

func TestMissingTimestampPrintsPadding(t *testing.T) {
	data := strings.NewReader(`level=info msg=hello`)
	w := &bytes.Buffer{}
	p := parser.NewParser(&config.Config{NoColor: true}, data, w)
	err := p.Start()
	require.NoError(t, err)
	assert.NotContains(t, w.String(), "0001-01-01")
	assert.Contains(t, w.String(), "msg=hello")
}

func TestTimestampLabelOnlyWhenParses(t *testing.T) {
	data := strings.NewReader(`time="2025-03-15T10:32:23Z" level=info msg=ok t=10s`)
	w := &bytes.Buffer{}
	p := parser.NewParser(&config.Config{NoColor: true}, data, w)
	err := p.Start()
	require.NoError(t, err)
	assert.Contains(t, w.String(), "t=10s")
}

func TestParseInvalidLogs(t *testing.T) {
	data := strings.NewReader(`
time="2025-03-15T10:32:23Z" level=info msg="loading"
time="`)
	w := &bytes.Buffer{}
	p := parser.NewParser(&config.Config{NoColor: true}, data, w)
	err := p.Start()
	require.NoError(t, err)
	assert.Contains(t, w.String(), "msg=loading")
	assert.Contains(t, w.String(), "time=")
}
