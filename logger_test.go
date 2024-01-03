// Copyright 2024 Paul Greenberg @greenpau
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
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLogger(t *testing.T) {
	for i, tc := range []struct {
		level   zapcore.Level
	}{
		{
			level:   zapcore.InfoLevel,
		},
		{
			level:   zapcore.DebugLevel,
		},
	} {
		logger := initLogger(tc.level)
		logger.Info("logging", zap.Int("testcase_id", i), zap.String("logger_log_level", tc.level.CapitalString()))
		t.Logf("PASS: Test %d", i)
	}
}