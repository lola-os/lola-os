// Package logging implements LOLA OS's structured logger. Every log line
// carries a level, a message, optional structured fields, and — per the
// branding guide — an icon and color hint. In "rich" format these render
// as ANSI-colored, icon-prefixed terminal output; in "json" format they
// are emitted as plain structured JSON lines (color/icon fields included
// as data, not escape codes) suitable for ingestion by other tools.
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelCritical
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

func (l Level) icon() string {
	switch l {
	case LevelDebug:
		return "🔄"
	case LevelInfo:
		return "🔵"
	case LevelWarn:
		return "⚠️"
	case LevelError:
		return "❌"
	case LevelCritical:
		return "🔴"
	default:
		return "•"
	}
}

func (l Level) colorFn() func(format string, a ...interface{}) string {
	switch l {
	case LevelDebug:
		return color.New(color.FgHiBlack).SprintfFunc()
	case LevelInfo:
		return color.New(color.FgBlue).SprintfFunc()
	case LevelWarn:
		return color.New(color.FgYellow).SprintfFunc()
	case LevelError:
		return color.New(color.FgRed).SprintfFunc()
	case LevelCritical:
		return color.New(color.FgHiRed, color.Bold).SprintfFunc()
	default:
		return color.New(color.Reset).SprintfFunc()
	}
}

func ParseLevel(s string) Level {
	switch s {
	case "debug":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "critical":
		return LevelCritical
	default:
		return LevelInfo
	}
}

// Entry is a single structured log line, also used as the wire format for
// `lola metrics` JSON-lines output and for SDK `stream_logs()` consumers.
type Entry struct {
	Time    time.Time              `json:"time"`
	Level   string                 `json:"level"`
	Icon    string                 `json:"icon"`
	Color   string                 `json:"color"`
	Message string                 `json:"message"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

// Logger writes Entry values to an underlying writer in either rich or
// json format. It is safe for concurrent use.
type Logger struct {
	mu       sync.Mutex
	out      io.Writer
	minLevel Level
	format   string // "rich" or "json"
}

// New constructs a Logger writing to w.
func New(w io.Writer, minLevel Level, format string) *Logger {
	if format != "json" {
		format = "rich"
	}
	return &Logger{out: w, minLevel: minLevel, format: format}
}

// Default returns a Logger writing to stderr (so stdout remains free for
// JSON-RPC responses) at info level in rich format.
func Default() *Logger {
	return New(os.Stderr, LevelInfo, "rich")
}

func (lg *Logger) log(level Level, msg string, fields map[string]interface{}) {
	if level < lg.minLevel {
		return
	}
	entry := Entry{
		Time:    time.Now().UTC(),
		Level:   level.String(),
		Icon:    level.icon(),
		Color:   colorName(level),
		Message: msg,
		Fields:  fields,
	}

	lg.mu.Lock()
	defer lg.mu.Unlock()

	if lg.format == "json" {
		enc := json.NewEncoder(lg.out)
		_ = enc.Encode(entry)
		return
	}

	cf := level.colorFn()
	prefix := cf("%s %-8s", entry.Icon, level.String())
	line := fmt.Sprintf("%s %s %s", entry.Time.Format("15:04:05"), prefix, msg)
	if len(fields) > 0 {
		b, _ := json.Marshal(fields)
		line += " " + string(b)
	}
	fmt.Fprintln(lg.out, line)
}

func colorName(l Level) string {
	switch l {
	case LevelInfo:
		return "blue"
	case LevelWarn:
		return "yellow"
	case LevelError, LevelCritical:
		return "red"
	default:
		return "gray"
	}
}

func (lg *Logger) Debug(msg string, fields map[string]interface{})    { lg.log(LevelDebug, msg, fields) }
func (lg *Logger) Info(msg string, fields map[string]interface{})     { lg.log(LevelInfo, msg, fields) }
func (lg *Logger) Warn(msg string, fields map[string]interface{})     { lg.log(LevelWarn, msg, fields) }
func (lg *Logger) Error(msg string, fields map[string]interface{})    { lg.log(LevelError, msg, fields) }
func (lg *Logger) Critical(msg string, fields map[string]interface{}) { lg.log(LevelCritical, msg, fields) }
