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

// Package tests - Cloud test tests
package tests

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/cloudtest/pkg/commands"
	"github.com/networkservicemesh/cloudtest/pkg/config"
	"github.com/networkservicemesh/cloudtest/pkg/utils"
)

func TestRestartRequest(t *testing.T) {
	logKeeper := utils.NewLogKeeper()
	defer logKeeper.Stop()

	testConfig := &config.CloudTestConfig{
		RetestConfig: config.RetestConfig{
			Patterns:     []string{"#Please_RETEST#"},
			RestartCount: 2,
		},
	}
	testConfig.Timeout = 3000

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	require.NoError(t, err)
	defer utils.ClearFolder(tmpDir, false)
	testConfig.ConfigRoot = tmpDir
	createProvider(testConfig, "a_provider")

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name: "simple",
		Source: config.ExecutionSource{
			Tags: []string{"request_restart"},
		},
		Timeout:     1500,
		PackageRoot: "./sample",
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	logrus.Info(err.Error())
	require.Contains(t, err.Error(), "there is failed tests 1")

	require.NotNil(t, report)

	rootSuite := report.Suites[0]

	require.Len(t, rootSuite.Suites, 1)
	require.Equal(t, 1, rootSuite.Suites[0].Failures)
	require.Len(t, rootSuite.Suites, 1)
	require.Len(t, rootSuite.Suites[0].Suites[0].TestCases, 2)

	logKeeper.CheckMessagesOrder(t, []string{
		"Starting TestRequestRestart",
		"Re schedule task TestRequestRestart reason: rerun-request",
		"Test TestRequestRestart retry count 2 exceed: err",
	})
	require.Equal(t, 2, logKeeper.MessageCount("Re schedule task TestRequestRestart reason: rerun-request"))
}

func TestRestartRetestDestroyCluster(t *testing.T) {
	logKeeper := utils.NewLogKeeper()
	defer logKeeper.Stop()

	testConfig := &config.CloudTestConfig{
		RetestConfig: config.RetestConfig{
			Patterns:       []string{"#Please_RETEST#"},
			RestartCount:   3,
			AllowedRetests: 1,
			WarmupTimeout:  0,
		},
	}
	testConfig.Timeout = 1000

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	require.NoError(t, err)
	defer utils.ClearFolder(tmpDir, false)
	testConfig.ConfigRoot = tmpDir
	p := createProvider(testConfig, "a_provider")
	p.Instances = 1
	p.RetryCount = 1

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name: "simple",
		Source: config.ExecutionSource{
			Tags: []string{"request_restart"},
		},
		Timeout:     1500,
		PackageRoot: "./sample",
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Contains(t, err.Error(), "there is failed tests 1")

	require.NotNil(t, report)

	rootSuite := report.Suites[0]
	require.Len(t, rootSuite.Suites, 2)

	require.Equal(t, 2, rootSuite.Suites[0].Tests)
	require.Equal(t, 0, rootSuite.Suites[0].Failures)
	require.Len(t, rootSuite.Suites[0].Suites[0].TestCases, 2)

	require.Equal(t, 1, rootSuite.Suites[1].Failures)
	require.Equal(t, 1, rootSuite.Suites[1].Tests)
	require.Len(t, rootSuite.Suites[1].TestCases, 1)

	logKeeper.CheckMessagesOrder(t, []string{
		"Starting TestRequestRestart",
		"Reached a limit of re-tests per cluster instance",
		"Destroying cluster",
		"Re schedule task TestRequestRestart reason: rerun-request",
		"Starting cluster ",
		"Starting TestRequestRestart",
		"Re schedule task TestRequestRestart reason: rerun-request",
		"Skip TestRequestRestart on a_provider-1: 1 of 1 required cluster(s) unavailable: [a_provider]",
	})
}

func TestRestartRequestRestartCluster(t *testing.T) {
	logKeeper := utils.NewLogKeeper()
	defer logKeeper.Stop()

	testConfig := &config.CloudTestConfig{
		RetestConfig: config.RetestConfig{
			Patterns:       []string{"#Please_RETEST#"},
			RestartCount:   3,
			AllowedRetests: 2,
			WarmupTimeout:  1,
		},
	}
	testConfig.Timeout = 1000

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	require.NoError(t, err)
	defer utils.ClearFolder(tmpDir, false)
	testConfig.ConfigRoot = tmpDir
	p := createProvider(testConfig, "a_provider")
	p.Instances = 1
	p.RetryCount = 10

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name: "simple",
		Source: config.ExecutionSource{
			Tags: []string{"request_restart"},
		},
		Timeout:     1500,
		PackageRoot: "./sample",
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Contains(t, err.Error(), "there is failed tests 1")

	require.NotNil(t, report)

	rootSuite := report.Suites[0]
	require.Len(t, rootSuite.Suites, 1)
	require.Equal(t, 1, rootSuite.Suites[0].Failures)
	require.Len(t, rootSuite.Suites, 1)
	require.Len(t, rootSuite.Suites[0].Suites[0].TestCases, 2)

	logKeeper.CheckMessagesOrder(t, []string{
		"Starting TestRequestRestart",
		"Reached a limit of re-tests per cluster instance",
		"Destroying cluster",
		"Starting cluster ",
		"Test TestRequestRestart retry count 3 exceed: err: failed to run go test",
	})
	require.Equal(t, 3, logKeeper.MessageCount("Re schedule task TestRequestRestart reason: rerun-request"))
}

func TestRestartRequestSkip(t *testing.T) {
	logKeeper := utils.NewLogKeeper()
	defer logKeeper.Stop()

	testConfig := &config.CloudTestConfig{
		RetestConfig: config.RetestConfig{
			Patterns:         []string{"#Please_RETEST#"},
			RestartCount:     2,
			RetestFailResult: "skip",
		},
	}
	testConfig.Timeout = 3000

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	require.NoError(t, err)
	defer utils.ClearFolder(tmpDir, false)
	testConfig.ConfigRoot = tmpDir
	createProvider(testConfig, "a_provider")

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name: "simple",
		Source: config.ExecutionSource{
			Tags: []string{"request_restart"},
		},
		Timeout:     1500,
		PackageRoot: "./sample",
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.NoError(t, err)

	require.NotNil(t, report)

	rootSuite := report.Suites[0]
	require.Len(t, rootSuite.Suites, 1)
	require.Equal(t, 0, rootSuite.Suites[0].Failures)
	require.Len(t, rootSuite.Suites, 1)
	require.Len(t, rootSuite.Suites[0].Suites[0].TestCases, 2)

	for _, tt := range rootSuite.Suites[0].TestCases {
		if tt.Name == "_TestRequestRestart" {
			require.Equal(t, "Test TestRequestRestart retry count 2 exceed: err: failed to run go test . -test.timeout 50m0s -count 1 --run \"^(TestRequestRestart)\\\\z\" --tags \"request_restart\" --test.v ExitCode: 1", tt.SkipMessage.Message)
		}
	}

	logKeeper.CheckMessagesOrder(t, []string{
		"Starting TestRequestRestart",
		"Re schedule task TestRequestRestart reason: rerun-request",
		"Test TestRequestRestart retry count 2 exceed: err",
		"Re schedule task TestRequestRestart reason: skipped",
	})
	require.Equal(t, 2, logKeeper.MessageCount("Re schedule task TestRequestRestart reason: rerun-request"))
}
