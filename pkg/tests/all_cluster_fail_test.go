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

package tests

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/cloudtest/pkg/commands"
	"github.com/networkservicemesh/cloudtest/pkg/config"
	"github.com/networkservicemesh/cloudtest/pkg/utils"
)

func TestClusterInstancesFailed(t *testing.T) {
	testConfig := config.NewCloudTestConfig()

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	require.NoError(t, err)
	defer utils.ClearFolder(tmpDir, false)
	testConfig.ConfigRoot = tmpDir
	createProvider(testConfig, "a_provider")
	failedP := createProvider(testConfig, "b_provider")
	failedP.Scripts["start"] = "echo starting\nexit 2"

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:        "simple",
		Timeout:     15,
		PackageRoot: "./sample",
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Contains(t, err.Error(), "there is failed tests 3")

	require.NotNil(t, report)

	rootSuite := report.Suites[0]

	require.Len(t, rootSuite.Suites, 2)

	require.Len(t, rootSuite.Suites[0].Suites, 2)
	require.Equal(t, 1, rootSuite.Suites[0].Failures)
	require.Equal(t, 6, rootSuite.Suites[0].Tests)

	require.Len(t, rootSuite.Suites[0].Suites[0].TestCases, 3)
	require.Len(t, rootSuite.Suites[0].Suites[1].TestCases, 3)

	require.Equal(t, 2, rootSuite.Suites[1].Failures)
	require.Equal(t, 2, rootSuite.Suites[1].Tests)
	require.Len(t, rootSuite.Suites[1].TestCases, 2)
}

func TestClusterInstancesFailedSpecificTestList(t *testing.T) {
	testConfig := &config.CloudTestConfig{}

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	require.NoError(t, err)
	defer utils.ClearFolder(tmpDir, false)
	testConfig.ConfigRoot = tmpDir
	createProvider(testConfig, "a_provider")

	testConfig.Executions = []*config.Execution{{
		Name:        "simple",
		Timeout:     2,
		PackageRoot: "./sample",
		Source: config.ExecutionSource{
			Tests: []string{"TestPass", "TestTimeout", "TestFail"},
		},
	}}

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Contains(t, err.Error(), "there is failed tests 2")

	require.NotNil(t, report)

	require.Len(t, report.Suites, 1)
	require.Equal(t, 2, report.Suites[0].Failures)
	require.Equal(t, 3, report.Suites[0].Tests)
}

func TestClusterInstancesOnFailGoRunner(t *testing.T) {
	testConfig := config.NewCloudTestConfig()

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	require.NoError(t, err)
	defer utils.ClearFolder(tmpDir, false)
	testConfig.ConfigRoot = tmpDir
	createProvider(testConfig, "a_provider")
	failedP := createProvider(testConfig, "b_provider")
	failedP.Scripts["start"] = "echo starting\nexit 2"

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:        "simple",
		Timeout:     15,
		PackageRoot: "./sample",
		OnFail:      `echo >>>Running on fail script<<<`,
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Contains(t, err.Error(), "there is failed tests 3")

	require.NotNil(t, report)

	rootSuite := report.Suites[0]

	require.Len(t, rootSuite.Suites, 2)

	require.Equal(t, 1, rootSuite.Suites[0].Failures)
	require.Equal(t, 6, rootSuite.Suites[0].Tests)
	require.Len(t, rootSuite.Suites[0].Suites[0].TestCases, 3)
	require.Len(t, rootSuite.Suites[0].Suites[1].TestCases, 3)

	require.Equal(t, 2, rootSuite.Suites[1].Failures)
	require.Equal(t, 2, rootSuite.Suites[1].Tests)
	require.Len(t, rootSuite.Suites[1].TestCases, 2)

	foundFailTest := false

	for _, execSuite := range rootSuite.Suites[0].Suites {
		if execSuite.Name == "a_provider" {
			for _, testCase := range execSuite.TestCases {
				if testCase.Name == "TestFail" {
					require.NotNil(t, testCase.Failure)
					require.Contains(t, testCase.Failure.Contents, ">>>Running on fail script<<<")
					foundFailTest = true
				} else {
					require.Nil(t, testCase.Failure)
				}
			}
		}
	}
	require.True(t, foundFailTest)
}

func TestClusterInstancesOnFailShellRunner(t *testing.T) {
	testConfig := config.NewCloudTestConfig()

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	require.NoError(t, err)
	defer utils.ClearFolder(tmpDir, false)
	testConfig.ConfigRoot = tmpDir
	createProvider(testConfig, "a_provider")
	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:    "pass",
		Timeout: 15,
		Kind:    "shell",
		Run:     "echo pass",
		OnFail:  `echo >>>Running on fail script<<<`,
	})
	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:    "fail",
		Timeout: 15,
		Kind:    "shell",
		Run:     "make_all_happy()",
		Env:     []string{"name=$(test-name)"},
		OnFail:  `echo >>>Running on fail script name=${name}, artifacts dir=${ARTIFACTS_DIR}<<< `,
	})
	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Contains(t, err.Error(), "there is failed tests 1")
	foundFailTest := false

	for _, executionSuite := range report.Suites[0].Suites {
		testCase := executionSuite.Suites[0].TestCases[0]
		if executionSuite.Name == "fail" {
			require.NotNil(t, testCase.Failure)
			require.Contains(t, testCase.Failure.Contents, ">>>Running on fail script name=OnFail, artifacts dir=")
			require.Contains(t, testCase.Failure.Contents, "fail<<<")
			foundFailTest = true
		} else {
			require.Nil(t, testCase.Failure)
		}
	}
	require.True(t, foundFailTest)
}

func TestClusterInstancesOnFailShellRunnerInterdomain(t *testing.T) {
	testConfig := config.NewCloudTestConfig()

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	require.NoError(t, err)
	defer utils.ClearFolder(tmpDir, false)
	testConfig.ConfigRoot = tmpDir
	ap := createProvider(testConfig, "a_provider")
	ap.Scripts["config"] = "echo ./.tests/config.a"
	bp := createProvider(testConfig, "b_provider")
	bp.Scripts["config"] = "echo ./.tests/config.b"
	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:            "pass",
		Timeout:         15,
		ClusterCount:    2,
		ClusterSelector: []string{"a_provider", "b_provider"},
		Kind:            "shell",
		Run:             "echo pass",
		OnFail:          `echo >>>Running on fail script with ${KUBECONFIG} <<<`,
	})
	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:            "fail",
		Timeout:         15,
		ClusterCount:    2,
		ClusterSelector: []string{"a_provider", "b_provider"},
		Kind:            "shell",
		Run:             "make_all_happy()",
		OnFail:          `echo >>>Running on fail script with ${KUBECONFIG} <<<`,
	})
	testConfig.Reporting.JUnitReportFile = JunitReport

	logKeeper := utils.NewLogKeeper()
	defer logKeeper.Stop()

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "there is failed tests 1")
	foundFailTest := false

	for _, suite := range report.Suites[0].Suites {
		ts := suite.Suites[0].TestCases[0]
		if suite.Name == "fail" {
			require.NotNil(t, ts.Failure)
			require.Contains(t, ts.Failure.Contents, ">>>Running on fail script with ./.tests/config.a <<<")
			require.Contains(t, ts.Failure.Contents, ">>>Running on fail script with ./.tests/config.b <<<")
			foundFailTest = true
		} else {
			require.Nil(t, ts.Failure)
		}
	}
	require.True(t, foundFailTest)

	require.Equal(t, 2, logKeeper.MessageCount("OnFail: running on fail script operations with"))
}
