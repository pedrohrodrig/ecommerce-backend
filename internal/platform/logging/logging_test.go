package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/config"
)

func TestNewWritesJSONInProduction(t *testing.T) {
	var out bytes.Buffer
	New(&out, config.Production, slog.LevelInfo).Info("hello", "answer", 42)

	var entry map[string]any
	if err := json.Unmarshal(out.Bytes(), &entry); err != nil {
		t.Fatalf("log line is not JSON: %q", out.String())
	}
	if entry["level"] != "INFO" || entry["msg"] != "hello" || entry["answer"] != float64(42) {
		t.Errorf("unexpected log entry: %v", entry)
	}
}

func TestNewWritesTextInDevelopment(t *testing.T) {
	var out bytes.Buffer
	New(&out, config.Development, slog.LevelInfo).Info("hello", "answer", 42)

	if got := out.String(); !strings.Contains(got, "level=INFO msg=hello answer=42") {
		t.Errorf("unexpected log line: %q", got)
	}
}

func TestNewFiltersBelowLevel(t *testing.T) {
	var out bytes.Buffer
	logger := New(&out, config.Development, slog.LevelWarn)
	logger.Info("noise")
	logger.Warn("kept")

	if got := out.String(); strings.Contains(got, "noise") || !strings.Contains(got, "kept") {
		t.Errorf("unexpected output for level warn: %q", got)
	}
}
