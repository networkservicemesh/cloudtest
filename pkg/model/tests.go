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

package model

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/cloudtest/pkg/config"
	"github.com/networkservicemesh/cloudtest/pkg/execmanager"
	"github.com/networkservicemesh/cloudtest/pkg/utils"
)

// Status - Test Execution status
type Status int8

const (
	// StatusAdded - test is added
	StatusAdded Status = iota // statusAdded - Just added
	// StatusSuccess - test is completed fine.
	StatusSuccess
	// StatusFailed - test is failed to be executed.
	StatusFailed
	// StatusTimeout - timeout during test execution.
	StatusTimeout
	// StatusSkipped - status if test is marked as skipped.
	StatusSkipped
	// StatusSkippedSinceNoClusters - status of test if not clusters of desired group are available.
	StatusSkippedSinceNoClusters
	// StatusRerunRequest - a test was requested its re-run
	StatusRerunRequest
)

// TestEntryExecution - represent one test execution.
type TestEntryExecution struct {
	OutputFile string // Output file name
	Retry      int    // Did we retry execution on this cluster.
	Status     Status // Execution status
}

//TestEntryKind - describes a testing way.
type TestEntryKind uint8

const (
	// GoTest - go test test
	GoTest TestEntryKind = iota
	// ShellTest - shell test.
	ShellTest
	// SuiteTest - go test suites
	SuiteTest
)

// TestEntry - represent one found test
type TestEntry struct {
	Name            string // Test name
	Tags            string // A list of tags
	Key             string // Unique key
	ExecutionConfig *config.Execution
	Suite           *Suite

	Executions []TestEntryExecution
	Duration   time.Duration
	Started    time.Time

	RunScript string

	Kind   TestEntryKind
	Status Status
	sync.Mutex
	SkipMessage         string
	ArtifactDirectories []string
}

// GetTestConfiguration - Return list of available tests by calling of gotest --list .* $root -tag "" and parsing of output.
func GetTestConfiguration(manager execmanager.ExecutionManager, root string, source config.ExecutionSource) (map[string]*TestEntry, error) {
	allTests, err1 := getTests(manager, root)
	if len(source.Tags) > 0 {
		tests, err := getTests(manager, root, source.Tags...)
		if err != nil {
			return nil, err
		}
		for key := range allTests {
			delete(tests, key)
		}
		return tests, nil
	} else if len(source.Tests) > 0 {
		result := map[string]*TestEntry{}
		var err error
		for _, n := range source.Tests {
			t := allTests[n]
			if t != nil {
				result[n] = t
			} else {
				msg := fmt.Sprintf("test %v not found", n)
				if err == nil {
					err = errors.New(msg)
				} else {
					err = errors.Wrap(err, msg)
				}
			}
		}
		return result, err
	}
	return allTests, err1
}

func getTests(manager execmanager.ExecutionManager, dir string, tags ...string) (map[string]*TestEntry, error) {
	gotestCmd := []string{"go", "test", ".", "--list", ".*"}
	tagsStr := strings.Join(tags, ",")
	if len(tagsStr) != 0 {
		gotestCmd = append(gotestCmd, "-tags", tagsStr)
	}

	result, err := utils.ExecRead(context.Background(), dir, gotestCmd)
	if err != nil {
		logrus.Errorf("Error getting list of tests: %v\nOutput: %v\nCmdLine: %v", err, result, gotestCmd)
		return nil, err
	}

	testResult := map[string]*TestEntry{}

	manager.AddLog("gotest", "find-tests", strings.Join(gotestCmd, " ")+"\n"+strings.Join(result, "\n"))
	for _, testLine := range result {
		if strings.ContainsAny(testLine, "\t") {
			special := strings.Split(testLine, "\t")
			if len(special) == 3 {
				// This is special case.
				continue
			}
		} else {
			testName := strings.TrimSpace(testLine)
			testResult[testName] = &TestEntry{
				Name: testName,
				Tags: tagsStr,
			}
		}
	}
	return testResult, nil
}
