// Copyright 2016 The Linux Foundation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	// TODO(xiekeyang): The adaptation should be checked for openSUSE
	// on non-x86_64 architectures. If encounting problem, the golang
	// version should be updated accordingly.
	"context"
	"os"

	"github.com/Sirupsen/logrus"
)

var (
	// G is an alias for GetLogger.
	//
	// We may want to define this locally to a package to get package tagged log
	// messages.
	G = GetLogger

	// LogEntry provides a public and standard logger instance.
	LogEntry = logrus.NewEntry(logrus.StandardLogger())
)

type (
	loggerKey struct{}
)

// WithLogger returns a new context with the provided logger. Use in
// combination with logger.WithField(s) for great effect.
func WithLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// GetLogger retrieves the current logger from the context. If no logger is
// available, the default logger is returned.
func GetLogger(ctx context.Context) *logrus.Entry {
	logger := ctx.Value(loggerKey{})

	if logger == nil {
		return LogEntry
	}

	return logger.(*logrus.Entry)
}

// EnableDebugMode enables a selectable debug mode.
func EnableDebugMode(debug bool) {
	if debug {
		LogEntry.Logger.Level = logrus.DebugLevel
	}
}

func init() {
	LogEntry.Logger.Out = os.Stderr
}
