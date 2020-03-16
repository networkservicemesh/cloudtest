// Copyright (c) 2020 Doc.ai, Inc and/or its affiliates.
//
// Copyright (c) 2019-2020 Cisco Systems, Inc and/or its affiliates.
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
	"time"

	"github.com/networkservicemesh/cloudtest/pkg/model"
	"github.com/networkservicemesh/cloudtest/pkg/shell"
	"github.com/networkservicemesh/cloudtest/pkg/utils"
)

type goTestRunner struct {
	test    *model.TestEntry
	cmdLine string
	envMgr  shell.EnvironmentManager
}

func (runner *goTestRunner) Run(timeoutCtx context.Context, env []string, writer *bufio.Writer) error {
	logger := func(s string) {}
	cmdEnv := append(runner.envMgr.GetProcessedEnv(), env...)
	_, err := utils.RunCommand(timeoutCtx, runner.cmdLine, runner.test.ExecutionConfig.PackageRoot,
		logger, writer, cmdEnv, nil, false)
	return err
}

func (runner *goTestRunner) GetCmdLine() string {
	return runner.cmdLine
}

// NewGoTestRunner - creates go test runner
func NewGoTestRunner(ids string, test *model.TestEntry, timeout time.Duration) TestRunner {
	cmdLine := fmt.Sprintf("go test . -test.timeout %v -count 1 --run \"^(%s)$\\\\z\" --tags \"%s\" --test.v",
		timeout, test.Name, test.Tags)

	envMgr := shell.NewEnvironmentManager()
	_ = envMgr.ProcessEnvironment(ids, "gotest", os.TempDir(), test.ExecutionConfig.Env, map[string]string{})
	return &goTestRunner{
		test:    test,
		cmdLine: cmdLine,
		envMgr:  envMgr,
	}
}
