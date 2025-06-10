package loggr

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer

	// Inject a test logger writing to the buffer
	logger := &LevelLogger{
		level:   LevelDebug, // will include Debug+
		appCode: "test-app",
		l:       newTestLogger(&buf),
	}

	// Capture a log
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Trace("trace should not appear") // won't appear with LevelDebug

	// Read lines
	logOutput := buf.String()
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")

	// Assertions
	assert.Len(t, lines, 2)
	assert.Contains(t, lines[0], "DEBUG")
	assert.Contains(t, lines[0], "debug message")
	assert.Contains(t, lines[1], "INFO")
	assert.Contains(t, lines[1], "info message")
	assert.NotContains(t, logOutput, "trace should not appear")
}

func newTestLogger(buf *bytes.Buffer) *log.Logger {
	return log.New(buf, "", 0)
}
