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
	"context"
	"net/http"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestFunctionExecutor(t *testing.T) {
	config := `
	lambda {
		name hello_world
		runtime python
		# python_executable {$HOME}/dev/go/src/github.com/greenpau/caddy-lambda/venv/bin/python
		python_executable python
		entrypoint assets/scripts/api/hello_world/app/index.py
		function handler
		workers 10
	}`
	for i, tc := range []struct {
		req *http.Request
		fex   FunctionExecutor
	}{
		{
			fex:   FunctionExecutor{
				Name: "foo",
			},
			req: newRequest(t, "GET", "/"),
		},
	} {
		d := caddyfile.NewTestDispenser(config)
		tc.fex.logger = initLogger(zapcore.DebugLevel)
		resp := newResponseWriter(tc.fex.logger)
		if err := tc.fex.UnmarshalCaddyfile(d); err != nil {
			t.Fatalf("unexpected UnmarshalCaddyfile() error: %v", err)
		}
		ctx := caddy.Context{Context: context.Background()}
		if err := tc.fex.Provision(ctx); err != nil {
			t.Fatalf("unexpected Provision() error: %v", err)
		}
		defer tc.fex.Cleanup()
		err := tc.fex.invoke(resp, tc.req)
		if err != nil {
			t.Fatalf("unexpected invoke() error: %v", err)
		}
		t.Logf("PASS: Test %d", i)
	}
}

func newRequest(t *testing.T, method, uri string) *http.Request {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	req.RequestURI = req.URL.RequestURI()
	return req
}

type responseWriter struct {
	body       []byte
	statusCode int
	header     http.Header
	logger		*zap.Logger
}

func newResponseWriter(logger *zap.Logger) *responseWriter {
	return &responseWriter{
		header: http.Header{},
		logger: logger,
	}
}

func (w *responseWriter) Header() http.Header {
	return w.header
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.logger.Debug("wrote response body", zap.ByteString("body", b))
	w.body = b
	return 0, nil
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.logger.Debug("wrote response header", zap.Int("status_code", statusCode))
}