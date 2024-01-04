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
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	caddy.RegisterModule(FunctionExecutor{})
}

// FunctionExecutor is a middleware which triggers execution of a function when
// it is invoked.
type FunctionExecutor struct {
	// Name stores the name associated with the function.
	Name string `json:"name,omitempty"`
	// Runtime stores the runtime of the function, e.g. python.
	Runtime string `json:"runtime,omitempty"`
	// EntrypointPath stores the path to the function's entrypoint, e.g. python script path.
	EntrypointPath string `json:"entrypoint_path,omitempty"`
	// EntrypointHandler stores the name of the function to invoke at the Entrypoint. e.g handler.
	EntrypointHandler string `json:"entrypoint_handler,omitempty"`
	// PythonExecutable stores the path to the python executable.
	PythonExecutable string `json:"python_executable,omitempty"`
	// MaxWorkersCount stores the max number of concurrent runtimes.
	MaxWorkersCount uint `json:"workers,omitempty"`
	// WorkerTimeout stores the maximum number of seconds a function would run.
	WorkerTimeout int `json:"worker_timeout,omitempty"`
	// If URIFilter is not empty, then only the plugin
	// intercepts only the pages matching the regular expression
	// in the filter
	URIFilter         string `json:"uri_filter,omitempty"`
	filterURIPattern  *regexp.Regexp
	logger            *zap.Logger
	workers           []*worker
	entrypointImport string
}

// CaddyModule returns the Caddy module information.
func (FunctionExecutor) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.lambda",
		New: func() caddy.Module { return new(FunctionExecutor) },
	}
}

// Provision sets up FunctionExecutor.
func (fex *FunctionExecutor) Provision(ctx caddy.Context) error {
	if fex.logger == nil {
		fex.logger = initLogger(zapcore.InfoLevel)
	}

	if fex.URIFilter != "" {
		p, err := regexp.CompilePOSIX(fex.URIFilter)
		if err != nil {
			return fmt.Errorf("failed to compile uri_filter %s: %s", fex.URIFilter, err)
		}
		fex.filterURIPattern = p
	}

	if fex.entrypointImport == "" {
		fex.entrypointImport = strings.ReplaceAll(fex.EntrypointPath, "/", ".")
		if strings.HasSuffix(fex.entrypointImport, ".py") {
			fex.entrypointImport = fex.entrypointImport[:len(fex.entrypointImport)-3]
		}
	}

	var workerID uint = 0
	if fex.WorkerTimeout < 1 {
		fex.WorkerTimeout = 60
	}
	timeout := time.Second * time.Duration(fex.WorkerTimeout)

	w, err := newWorker(workerID, fex.PythonExecutable, []string{"-u", "-q", "-i"}, timeout, fex.logger)
	if err != nil {
		return fmt.Errorf("failed starting lambda worker %d %s: %s", workerID, fex.Name, err)
	}
	fex.workers = append(fex.workers, w)

	fex.logger.Info(
		"started lambda runtime",
		zap.String("lambda_name", fex.Name),
		zap.Uint("worker_id", workerID),
		zap.Int("worker_pid", w.getProcessPid()),
		zap.Int("worker_timeout", fex.WorkerTimeout),
	)
	return nil
}

func (fex FunctionExecutor) ServeHTTP(resp http.ResponseWriter, req *http.Request, next caddyhttp.Handler) error {
	return fex.invoke(resp, req)
}

// Cleanup implements caddy.CleanerUpper and terminates running processes.
func (fex *FunctionExecutor) Cleanup() error {
	fex.logger.Info(
		"cleaning up plugin",
		zap.String("plugin_name", pluginName),
		zap.String("lambda_name", fex.Name),
	)

	for _, w := range fex.workers {
		if err := w.terminate(); err != nil {
			fex.logger.Warn(
				"failed shutting down lambda runtime",
				zap.String("plugin_name", pluginName),
				zap.String("lambda_name", fex.Name),
				zap.Uint("worker_id", w.ID),
				zap.Int("worker_pid", w.Pid),
				zap.Error(err),
			)
			continue
		}
		fex.logger.Info(
			"completed shutdown of lambda runtime",
			zap.String("plugin_name", pluginName),
			zap.String("lambda_name", fex.Name),
			zap.Uint("worker_id", w.ID),
			zap.Int("worker_pid", w.Pid),
		)
	}
	return nil
}

// Interface guard
var _ caddyhttp.MiddlewareHandler = (*FunctionExecutor)(nil)
