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
	"net/http"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestFunctionExecutor(t *testing.T) {
	for i, tc := range []struct {
		req *http.Request
		resp http.ResponseWriter
		fex   FunctionExecutor
	}{
		{
			fex:   FunctionExecutor{
				Name: "foo",
			},
			resp: newResponseWriter(),
			req: newRequest(t, "GET", "/"),
		},
	} {
		tc.fex.logger = initLogger(zapcore.DebugLevel)
		err := tc.fex.invoke(tc.resp, tc.req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
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
}

func newResponseWriter() *responseWriter {
	return &responseWriter{
		header: http.Header{},
	}
}

func (w *responseWriter) Header() http.Header {
	return w.header
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = b
	return 0, nil
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}