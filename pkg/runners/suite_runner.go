// Copyright (c) 2020 Doc.ai and/or its affiliates.
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

package runners

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/edwarnicke/exechelper"

	"github.com/networkservicemesh/cloudtest/pkg/model"
	"github.com/networkservicemesh/cloudtest/pkg/shell"
)

type SuiteRunner struct {
	cmd        string
	envManager shell.EnvironmentManager
	test       *model.TestEntry
}

func (s *SuiteRunner) Run(ctx context.Context, envs []string, writer *bufio.Writer) error {
	envs = append(envs, s.envManager.GetProcessedEnv()...)
	err := exechelper.Run(s.cmd,
		exechelper.WithStdout(writer),
		exechelper.WithStderr(writer),
		exechelper.WithContext(ctx),
		exechelper.WithDir(s.test.ExecutionConfig.PackageRoot))
	exechelper.WithEnvirons(envs...)
	return err
}

func (s *SuiteRunner) GetCmdLine() string {
	return s.cmd
}

var _ TestRunner = (*SuiteRunner)(nil)

func NewSuiteRunner(ids string, test *model.TestEntry, timeout time.Duration) *SuiteRunner {
	pattern := strings.Join(test.Suite.Tests, "|")
	cmdLine := fmt.Sprintf("go test . -testify.m \"%v\" -test.timeout %v -count 1 --run \"^(%s)$\\\\z\" --tags \"%s\" --test.v",
		pattern, timeout, test.Suite.Name, test.Tags)

	envMgr := shell.NewEnvironmentManager()
	_ = envMgr.ProcessEnvironment(ids, "gotest", os.TempDir(), test.ExecutionConfig.Env, map[string]string{})
	return &SuiteRunner{
		test:       test,
		cmd:        cmdLine,
		envManager: envMgr,
	}
}
