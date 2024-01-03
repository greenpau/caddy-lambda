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
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"

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
	// If URIFilter is not empty, then only the plugin
	// intercepts only the pages matching the regular expression
	// in the filter
	URIFilter        string `json:"uri_filter,omitempty"`
	filterURIPattern *regexp.Regexp
	logger           *zap.Logger
	cmd				*exec.Cmd
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
	// fex.logger = ctx.Logger(fex)
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

	// Run server
	var stdin, stdout, stderr bytes.Buffer
	cmd := exec.Command(fex.PythonExecutable)
	cmd.Stdin = &stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed starting lambda %s: %s", fex.Name, err)
	}

	fex.logger.Info(
		"started lambda runtime", 
		zap.String("lambda_name", fex.Name),
		zap.Int("pid", cmd.Process.Pid),
	)
	fex.cmd = cmd
	return nil
}

func (fex FunctionExecutor) ServeHTTP(resp http.ResponseWriter, req *http.Request, next caddyhttp.Handler) error {
	return fex.invoke(resp, req)
}

// Cleanup implements caddy.CleanerUpper and terminates running processes. 
func (fex *FunctionExecutor) Cleanup() error {
	if fex.cmd == nil {
		return nil
	}
	if fex.cmd.Process == nil {
		return nil
	}

	fex.logger.Info(
		"cleaning up lambda plugin", 
		zap.String("lambda_name", fex.Name),
		zap.Int("pid", fex.cmd.Process.Pid),
	)

	err := fex.cmd.Process.Kill()
	fex.logger.Info(
		"shutting down lambda runtime", 
		zap.String("lambda_name", fex.Name),
		zap.Error(err),
	)
	if err != nil {
		return err
	}
	err = fex.cmd.Wait()

	if err == nil {
		return err
	}
	if err.Error() == "signal: killed" {
		fex.logger.Info(
			"completed shutdown of lambda runtime", 
			zap.String("lambda_name", fex.Name),
		)
		return nil
	}
	fex.logger.Info(
		"completed shutdown of lambda runtime", 
		zap.String("lambda_name", fex.Name),
		zap.Error(err),
	)
	return err
}

// Interface guard
var _ caddyhttp.MiddlewareHandler = (*FunctionExecutor)(nil)
