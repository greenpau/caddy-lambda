// Copyright 2022 Paul Greenberg greenpau@outlook.com
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

package lambda

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// initDebugLogger returns an instance of debug-level logger.
func initDebugLogger() *zap.Logger {
	return initLogger(zapcore.DebugLevel)
}

// initInfoLogger returns an instance of info-level logger.
func initInfoLogger() *zap.Logger {
	return initLogger(zapcore.InfoLevel)
}

// initLogger returns an instance of logger.
func initLogger(level zapcore.Level) *zap.Logger {
	logAtom := zap.NewAtomicLevel()
	logAtom.SetLevel(level)
	logEncoderConfig := zap.NewProductionEncoderConfig()
	logEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logEncoderConfig.TimeKey = "time"
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(logEncoderConfig),
		zapcore.Lock(os.Stdout),
		logAtom,
	))
	return logger
}