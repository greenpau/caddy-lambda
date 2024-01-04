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

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/go-cmp/cmp"
)

func TestParseCaddyfile(t *testing.T) {
	testcases := []struct {
		name      string
		d         *caddyfile.Dispenser
		want      string
		shouldErr bool
		err       error
	}{
		{
			name: "test python runtime",
			d: caddyfile.NewTestDispenser(`
				lambda {
					name hello_world
					runtime python
					python_executable {$HOME}/dev/go/src/github.com/greenpau/caddy-lambda/venv/bin/python
			  		entrypoint assets/scripts/api/hello_world/app/index.py
					function handler
					workers 1
				}`),
			want: `{
			    "foo": "bar"
			}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			fex := &FunctionExecutor{}
			fex.logger = initDebugLogger()
			err := fex.UnmarshalCaddyfile(tc.d)
			if err != nil {
				if !tc.shouldErr {
					t.Fatalf("expected success, got: %v", err)
				}
				if diff := cmp.Diff(err.Error(), tc.err.Error()); diff != "" {
					t.Fatalf("unexpected error: %v, want: %v", err, tc.err)
				}
				return
			}
			if tc.shouldErr {
				t.Fatalf("unexpected success, want: %v", tc.err)
			}

			// fullCfg := unpack(t, string(app.(httpcaddyfile.).Value))
			// cfg := fullCfg["config"].(map[string]interface{})

			// got := make(map[string]interface{})
			// for _, k := range []string{"credentials"} {
			// 	got[k] = cfg[k].(map[string]interface{})
			// }

			// want := unpack(t, tc.want)

			// if diff := cmp.Diff(want, got); diff != "" {
			// 	t.Errorf("parseCaddyfileAppConfig() mismatch (-want +got):\n%s", diff)
			// }
		})
	}

}