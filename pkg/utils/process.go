// Copyright (c) 2019 Cisco Systems, Inc and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
)

//ProcWrapper - A simple process wrapper
type ProcWrapper struct {
	Cmd    *exec.Cmd
	cancel context.CancelFunc
	Stdout io.ReadCloser
	Stderr io.ReadCloser
}

// ExitCode - wait for completion and return exit code
func (w *ProcWrapper) ExitCode() int {
	err := w.Cmd.Wait()
	if err != nil {
		e, ok := err.(*exec.ExitError)
		if ok {
			return e.ExitCode()
		}
		logrus.Errorf("Error during waiting for process exit code: %v %v", w.Cmd.Args, err)
		return -1
	}
	return w.Cmd.ProcessState.ExitCode()
}

// ExecRead - execute command and return output as result, stderr is ignored.
func ExecRead(ctx context.Context, dir string, args []string) ([]string, error) {
	proc, err := ExecProc(ctx, dir, args, nil)
	if err != nil {
		return nil, err
	}
	output := []string{}
	reader := bufio.NewReader(proc.Stdout)
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		output = append(output, strings.TrimSpace(s))
	}
	err = proc.Cmd.Wait()
	if err != nil {
		return output, err
	}
	return output, nil
}

// ExecProc - execute shell command and return ProcWrapper
func ExecProc(ctx context.Context, dir string, args, env []string) (*ProcWrapper, error) {
	if len(args) == 0 {
		return &ProcWrapper{}, errors.New("missing command to run")
	}

	ctx, cancel := context.WithCancel(ctx)
	p := &ProcWrapper{
		Cmd:    exec.CommandContext(ctx, args[0], args[1:]...),
		cancel: cancel,
	}
	p.Cmd.Dir = dir
	if env != nil {
		p.Cmd.Env = append(os.Environ(), env...)
	}
	var err error
	p.Stdout, err = p.Cmd.StdoutPipe()
	if err != nil {
		return p, err
	}
	p.Stderr, err = p.Cmd.StderrPipe()
	if err != nil {
		return p, err
	}
	err = p.Cmd.Start()
	return p, err
}
