package cmd

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func levelCmd(t *testing.T, value string) *cobra.Command {
	t.Helper()
	c := &cobra.Command{}
	c.Flags().String("log-level", "info", "")
	if value != "" {
		if err := c.Flags().Set("log-level", value); err != nil {
			t.Fatalf("set flag: %v", err)
		}
	}
	return c
}

func TestLogLevelFlag(t *testing.T) {
	tests := []struct {
		value string
		want  slog.Level
	}{
		{"", slog.LevelInfo}, // the default
		{"error", slog.LevelError},
		{"warn", slog.LevelWarn},
		{"info", slog.LevelInfo},
		{"debug", slog.LevelDebug},
		{"Debug", slog.LevelDebug},
	}
	for _, tt := range tests {
		got, err := logLevel(levelCmd(t, tt.value))
		if err != nil || got != tt.want {
			t.Errorf("--log-level=%q: got %v, %v; want %v", tt.value, got, err, tt.want)
		}
	}
}

func TestNewLoggerFiltersByFlag(t *testing.T) {
	var buf bytes.Buffer
	logger, err := newLogger(levelCmd(t, "info"), &buf)
	if err != nil {
		t.Fatalf("newLogger: %v", err)
	}
	logger.Debug("not shown")
	logger.Info("shown")
	if buf.String() != "shown\n" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestRootRejectsUnknownLogLevel(t *testing.T) {
	if f := rootCmd.Flags().Lookup("log-level"); f == nil || f.DefValue != "info" {
		t.Fatalf("expected a --log-level flag defaulting to info, got %+v", f)
	}

	rootCmd.SetArgs([]string{"--browser=false", "--port=1", "--log-level=loud"})
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		_ = rootCmd.Flags().Set("log-level", "info")
		_ = rootCmd.Flags().Set("browser", "true")
		_ = rootCmd.Flags().Set("port", "6419")
	})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected an error for an unknown log level")
	}
	msg := err.Error()
	for _, want := range []string{"--log-level", `"loud"`, "error, warn, info, debug"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q should mention %s", msg, want)
		}
	}
}
