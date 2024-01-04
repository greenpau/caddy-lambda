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
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

const (
	pluginName = "lambda"
)

func init() {
	httpcaddyfile.RegisterHandlerDirective(pluginName, parseCaddyfile)
}

// parseCaddyfile sets up a handler for function execution
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var fex FunctionExecutor
	fex.logger = initDebugLogger()
	err := fex.UnmarshalCaddyfile(h.Dispenser)
	return fex, err
}

func ensureArgsCount(d *caddyfile.Dispenser, args []string, count int) error {
	if len(args) != count {
		return d.Errf("too many args %q, expected %d", args, count)
	}
	return nil
}

func ensureArgUint(d *caddyfile.Dispenser, name, arg string) (uint, error) {
	n, err := strconv.Atoi(arg)
    if err != nil {
		return 0, d.Errf("failed to convert %s %s: %v", name, arg, err)
    }
	ns := strconv.Itoa(n)
	if ns != arg {
		return 0, d.Errf("failed to convert %s %s, resolved %s", name, arg, ns)
	}
	if n < 0 {
		return 0, d.Errf("%s %s must be greater or equal to zero", name, arg)
	}

	return uint(n), nil
}

// UnmarshalCaddyfile sets up the handler from Caddyfile tokens. Syntax:
//
//	lambda [<matcher>] {
//      name <name>
//      runtime <name>
//      entrypoint <path>
//      function <name>
//	}
func (fex *FunctionExecutor) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		args := d.RemainingArgs()
		if len(args) > 0 {
			return d.ArgErr()
		}

		for d.NextBlock(0) {
			switch d.Val() {
			case "name":
				args = d.RemainingArgs()
				err := ensureArgsCount(d, args, 1)
				if err != nil {
					return err
				}				
				fex.Name = args[0]
			case "runtime":
				args = d.RemainingArgs()
				err := ensureArgsCount(d, args, 1)
				if err != nil {
					return err
				}				
				fex.Runtime = args[0]
			case "python_executable":
				args = d.RemainingArgs()
				err := ensureArgsCount(d, args, 1)
				if err != nil {
					return err
				}				
				fex.PythonExecutable = args[0]
			case "entrypoint":
				args = d.RemainingArgs()
				err := ensureArgsCount(d, args, 1)
				if err != nil {
					return err
				}				
				fex.EntrypointPath = args[0]
			case "function":
				args = d.RemainingArgs()
				err := ensureArgsCount(d, args, 1)
				if err != nil {
					return err
				}				
				fex.EntrypointHandler = args[0]
			case "workers":
				args = d.RemainingArgs()
				err := ensureArgsCount(d, args, 1)
				if err != nil {
					return err
				}
				count, err := ensureArgUint(d, "workers", args[0])
				if err != nil {
					return err
				}
				fex.MaxWorkersCount = count
			default:
				return d.Errf("unsupported %s directive %q", pluginName, d.Val())
			}
		}
	}

	switch fex.Runtime {
	case "python":
		if fex.Name == "" {
			return d.Err("lambda name is not set")
		}
		if fex.EntrypointPath == "" {
			return d.Errf("%s lambda %s runtime entrypoint path is not set", fex.Name, fex.Runtime)
		}
		if fex.EntrypointHandler == "" {
			return d.Errf("%s lambda %s runtime entrypoint function is not set", fex.Name, fex.Runtime)
		}
		if fex.PythonExecutable == "" {
			fex.PythonExecutable = "python"
		}
		if fex.MaxWorkersCount == 0 {
			fex.MaxWorkersCount = 1
		}
		fex.logger.Debug(
			"configured lambda function",
			zap.String("name", fex.Name),
			zap.String("runtime", fex.Runtime),
			zap.String("python_executable", fex.PythonExecutable),
			zap.String("entrypoint", fex.EntrypointPath),
			zap.String("function", fex.EntrypointHandler),
			zap.Uint("workers", fex.MaxWorkersCount),
		)
	default:
		return d.Errf("lambda runtime is not set")
	}

	return nil
}