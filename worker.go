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
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type worker struct {
	mu             sync.RWMutex
	ID             uint
	InUse          bool
	Terminated     bool
	Cmd            *exec.Cmd
	Pid            int
	stdin          io.WriteCloser
	stdout         io.ReadCloser
	stderr         io.ReadCloser
	timeout        time.Duration
	importComplete bool
	logger         *zap.Logger
}

func newWorker(id uint, binPath string, args []string, timeout time.Duration, logger *zap.Logger) (*worker, error) {
	w := &worker{
		ID:     id,
		logger: logger,
	}

	cmd := exec.Command(binPath, args...)

	cmdStdin, cmdStdinErr := cmd.StdinPipe()
	if cmdStdinErr != nil {
		return nil, cmdStdinErr
	}
	cmdStdout, cmdStdoutErr := cmd.StdoutPipe()
	if cmdStdoutErr != nil {
		return nil, cmdStdoutErr
	}
	cmdStderr, cmdStderrErr := cmd.StderrPipe()
	if cmdStderrErr != nil {
		return nil, cmdStderrErr
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	w.Cmd = cmd
	w.Pid = cmd.Process.Pid
	w.stdin = cmdStdin
	w.stdout = cmdStdout
	w.stderr = cmdStderr
	w.timeout = timeout
	return w, nil
}

// getProcessPid returns process id of the worker.
func (w *worker) getProcessPid() int {
	return w.Pid
}

// terminate shuts down the worker.
func (w *worker) terminate() error {
	w.Terminated = true
	if w.Cmd == nil {
		return nil
	}
	if w.Cmd.Process == nil {
		return nil
	}
	err := w.Cmd.Process.Kill()
	if err != nil {
		return err
	}

	err = w.Cmd.Wait()
	if err == nil {
		return nil
	}
	if err.Error() == "signal: killed" {
		return nil
	}
	return err
}

func readPipe(ch chan string, stopWord string, timeout time.Duration) ([]string, bool) {
	var lines []string
	for {
		select {
		case line, ok := <-ch:
			if !ok {
				return lines, false
			}
			lines = append(lines, line)
			if strings.Contains(line, stopWord) {
				return lines, false
			}
		case <-time.After(timeout):
			return lines, true
		}
	}
}

func pipeListener(pipe io.Reader) chan string {
	ch := make(chan string)
	go func(ch chan string) {
		defer close(ch)
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			ch <- scanner.Text()
		}
	}(ch)
	return ch
}

func parseStatusCode(s string) (int, error) {
	s = strings.ReplaceAll(s, "CMD_STATUS_CODE=", "")
	s = strings.ReplaceAll(s, ";", "")
	n, err := strconv.Atoi(s)
	if err == nil {
		return n, nil
	}
	return 0, fmt.Errorf("failed to parse integer from input string: %s", s)
}

func (w *worker) handle(importedPath, handlerName string, data map[string]interface{}) (int, []byte, error) {
	w.mu.Lock()
	w.InUse = true
	defer func() {
		w.mu.Unlock()
		w.InUse = false
	}()

	if !w.importComplete {
		io.WriteString(w.stdin, "from "+importedPath+" import *")
		io.WriteString(w.stdin, "\n")
		io.WriteString(w.stdin, "import json")
		io.WriteString(w.stdin, "\n")
		w.importComplete = true
	}

	// Marshal the map into a JSON byte slice
	encodedData, err := json.Marshal(data)
	if err != nil {
		return http.StatusBadRequest, []byte(http.StatusText(http.StatusBadRequest)), nil
	}

	// Convert the byte slice to a JSON string
	requestID := data["request_id"].(string)
	stdout := pipeListener(w.stdout)
	io.WriteString(w.stdin, `resp = handler(` + string(encodedData) + `)`)
	io.WriteString(w.stdin, "\n")
	io.WriteString(w.stdin, `print("CMD_OUTPUT_START=`+requestID+`;")`)
	io.WriteString(w.stdin, "\n")
	io.WriteString(w.stdin, `print(f"CMD_STATUS_CODE={resp['status_code']};")`)
	io.WriteString(w.stdin, "\n")
	io.WriteString(w.stdin, `print(f"CMD_OUTPUT_BODY={resp['body']}")`)
	io.WriteString(w.stdin, "\n")
	io.WriteString(w.stdin, `print(f"CMD_OUTPUT_END=`+requestID+`;")`)
	io.WriteString(w.stdin, "\n")
	lines, timedOut := readPipe(stdout, "CMD_OUTPUT_END=", w.timeout)
	recordingOn := false
	statusCode := 200
	stdoutOutput := []string{}
	for _, line := range lines {
		if !recordingOn {
			if strings.HasPrefix(line, "CMD_OUTPUT_START=") {
				if strings.HasPrefix(line, "CMD_OUTPUT_START="+requestID+";") {
					recordingOn = true
				}
			}
			continue
		}
		if strings.HasPrefix(line, "CMD_STATUS_CODE=") {
			code, err := parseStatusCode(line)
			if err != nil {
				w.logger.Warn(
					"encountered error",
					zap.String("request_id", requestID),
					zap.Error(err),
				)
			} else {
				statusCode = code
			}
			continue
		}
		if strings.HasPrefix(line, "CMD_OUTPUT_BODY=") {
			stdoutOutput = append(stdoutOutput, strings.ReplaceAll(line, "CMD_OUTPUT_BODY=", ""))
			continue
		}

		if strings.HasPrefix(line, "CMD_OUTPUT_END=") {
			if strings.HasPrefix(line, "CMD_OUTPUT_END="+requestID+";") {
				recordingOn = false
				continue
			}
		}
		stdoutOutput = append(stdoutOutput, "Y"+line)
	}

	output := strings.Join(stdoutOutput, "\n")
	if timedOut {
		return http.StatusRequestTimeout, []byte(http.StatusText(http.StatusRequestTimeout)), nil
	}

	return statusCode, []byte(output), nil
}
