// Copyright (c) 2021 Doc.ai and/or its affiliates.
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

package suites

import (
	"os"
	"strings"

	"github.com/networkservicemesh/cloudtest/pkg/execmanager"
	"github.com/networkservicemesh/cloudtest/pkg/model"
	"github.com/networkservicemesh/cloudtest/pkg/suites/parse"
	"github.com/networkservicemesh/cloudtest/pkg/suites/testentry"
)

const (
	setupSuite = "SetupSuite"
)

// SkipSuite returns list of model.TestEntry for the skipped go suite
func SkipSuite(suite *model.TestEntry) (tests []*model.TestEntry) {
	for _, testName := range suite.Suite.Tests {
		tests = append(tests, &model.TestEntry{
			Name:            testName,
			ExecutionConfig: suite.ExecutionConfig,
			Started:         suite.Started,
			RunScript:       suite.RunScript,
			Kind:            model.GoTestKind,
			Status:          suite.Status,
		})
	}

	return tests
}

// SplitSuite returns list of model.TestEntry for the passed/failed go suite
func SplitSuite(
	suite *model.TestEntry,
	manager execmanager.ExecutionManager,
	clusterTaskID string,
) (tests []*model.TestEntry, err error) {
	suiteTests := make([]string, len(suite.Suite.Tests)+1)
	suiteTests[0] = setupSuite
	copy(suiteTests[1:], suite.Suite.Tests)

	builders := make(map[string]*testentry.Builder)
	for _, testName := range suiteTests {
		builders[testName] = testentry.NewBuilder(testName, suite, manager, clusterTaskID)
	}

	if err = splitExecutions(suite, builders); err != nil {
		return nil, err
	}

	setup := builders[setupSuite].Build()
	delete(builders, setupSuite)

	allSkip := true
	for _, builder := range builders {
		testEntry := builder.Build()
		allSkip = allSkip && (testEntry.Status == model.StatusSkipped)
		tests = append(tests, builder.Build())
	}

	if allSkip && setup.Status == model.StatusFailed {
		tests = append([]*model.TestEntry{setup}, tests...)
	}

	return tests, nil
}

func splitExecutions(suite *model.TestEntry, builders map[string]*testentry.Builder) error {
	for _, execution := range suite.Executions {
		file, err := os.Open(execution.OutputFile)
		if err != nil {
			return err
		}

		run := make(map[string]struct{})
		for event := range parse.Events(file) {
			if event.Err != nil {
				return event.Err
			}

			var testName string
			switch {
			case event.Test == "" || strings.HasPrefix(suite.Name, event.Test):
				testName = setupSuite
			default:
				testName = event.TestName()
			}

			if err = event.Process(builders[testName]); err != nil {
				return err
			}
			run[testName] = struct{}{}
		}

		for testName, builder := range builders {
			if _, ok := run[testName]; !ok {
				if err = builder.ProcessRunEvent(&parse.TestEvent{Time: suite.Started}); err != nil {
					return err
				}
				if err = builder.ProcessSkipEvent(&parse.TestEvent{
					Time:   suite.Started,
					Output: "Suite setup failed",
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
