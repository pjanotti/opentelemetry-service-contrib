// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package logzioexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/logzioexporter"

import (
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// hclog2ZapLogger implements Hashicorp's hclog.Logger interface using Uber's zap.Logger. It's a workaround for plugin
// system. go-plugin doesn't support other logger than hclog. This logger implements only methods used by the go-plugin.
type hclog2ZapLogger struct {
	Zap  *zap.Logger
	name string
}

func (*hclog2ZapLogger) Log(hclog.Level, string, ...any) {}

func (*hclog2ZapLogger) ImpliedArgs() []any {
	return nil
}

func (l *hclog2ZapLogger) Name() string {
	return l.name
}

func (*hclog2ZapLogger) StandardWriter(*hclog.StandardLoggerOptions) io.Writer {
	return nil
}

// Trace implementation.
func (*hclog2ZapLogger) Trace(string, ...any) {}

// Debug implementation.
func (l *hclog2ZapLogger) Debug(msg string, args ...any) {
	l.Zap.Debug(msg, argsToFields(args...)...)
}

// Info implementation.
func (l *hclog2ZapLogger) Info(msg string, args ...any) {
	l.Zap.Info(msg, argsToFields(args...)...)
}

// Warn implementation.
func (l *hclog2ZapLogger) Warn(msg string, args ...any) {
	l.Zap.Warn(msg, argsToFields(args...)...)
}

// Error implementation.
func (l *hclog2ZapLogger) Error(msg string, args ...any) {
	l.Zap.Error(msg, argsToFields(args...)...)
}

// IsTrace implementation.
func (*hclog2ZapLogger) IsTrace() bool { return false }

// IsDebug implementation.
func (*hclog2ZapLogger) IsDebug() bool { return false }

// IsInfo implementation.
func (*hclog2ZapLogger) IsInfo() bool { return false }

// IsWarn implementation.
func (*hclog2ZapLogger) IsWarn() bool { return false }

// IsError implementation.
func (*hclog2ZapLogger) IsError() bool { return false }

// With implementation.
func (l *hclog2ZapLogger) With(args ...any) hclog.Logger {
	return &hclog2ZapLogger{Zap: l.Zap.With(argsToFields(args...)...)}
}

// Named implementation.
func (l *hclog2ZapLogger) Named(name string) hclog.Logger {
	return &hclog2ZapLogger{Zap: l.Zap.Named(name)}
}

// ResetNamed implementation.
func (*hclog2ZapLogger) ResetNamed(string) hclog.Logger {
	// no need to implement that as go-plugin doesn't use this method.
	return &hclog2ZapLogger{}
}

// SetLevel implementation.
func (*hclog2ZapLogger) SetLevel(hclog.Level) {
	// no need to implement that as go-plugin doesn't use this method.
}

// GetLevel implementation.
func (*hclog2ZapLogger) GetLevel() hclog.Level {
	// no need to implement that as go-plugin doesn't use this method.
	return hclog.NoLevel
}

// StandardLogger implementation.
func (*hclog2ZapLogger) StandardLogger(*hclog.StandardLoggerOptions) *log.Logger {
	// no need to implement that as go-plugin doesn't use this method.
	return log.New(io.Discard, "", 0)
}

func argsToFields(args ...any) []zapcore.Field {
	var fields []zapcore.Field
	for i := 0; i < len(args); i += 2 {
		fields = append(fields, zap.String(args[i].(string), fmt.Sprintf("%v", args[i+1])))
	}

	return fields
}
