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

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (fex *FunctionExecutor) invoke(resp http.ResponseWriter, req *http.Request) error {
	if fex.filterURIPattern != nil {
		if !fex.filterURIPattern.MatchString(req.RequestURI) {
			return nil
		}
	}
	var requestID string
	rawRequestID := caddyhttp.GetVar(req.Context(), "request_id")
	if rawRequestID == nil {
		requestID = uuid.New().String()
		caddyhttp.SetVar(req.Context(), "request_id", requestID)
	} else {
		requestID = rawRequestID.(string)
	}

	// Extract cookies
	cookies := req.Cookies()

	// Extract query parameters
	queryParams := make(map[string]interface{})
	queryValues := req.URL.Query()
	for k, v := range queryValues {
		if len(v) == 1 {
			queryParams[k] = v[0]
		} else {
			queryParams[k] = v
		}
	}

	// Extract headers
	reqHeaders := make(map[string]interface{})
	if req.Header != nil {
		for k, v := range req.Header {
			if k == "Cookie" || k == "Set-Cookie" {
				continue
			}
			if len(v) == 1 {
				reqHeaders[k] = v[0]
			} else {
				reqHeaders[k] = v
			}
		}
	}

	fex.logger.Debug(
		"invoked lambda function",
		zap.String("lambda_name", fex.Name),
		zap.String("request_id", requestID),
		zap.String("method", req.Method),
		zap.String("proto", req.Proto),
		zap.String("host", req.Host),
		zap.String("uri", req.RequestURI),
		zap.String("remote_addr_port", req.RemoteAddr),
		zap.Int64("content_length", req.ContentLength),
		zap.Int("cookie_count", len(cookies)),
		zap.String("user_agent", req.UserAgent()),
		zap.String("referer", req.Referer()),
		zap.Any("cookies", cookies),
		zap.Any("query_params", queryParams),
		zap.Any("headers", reqHeaders),
	)

	resp.WriteHeader(200)
	resp.Write([]byte(`FOO`))

	return nil
}