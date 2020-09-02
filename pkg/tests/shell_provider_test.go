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
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/cloudtest/pkg/commands"
	"github.com/networkservicemesh/cloudtest/pkg/config"
	"github.com/networkservicemesh/cloudtest/pkg/utils"
)

const (
	JunitReport = "reporting/junit.xml"
)

func TestShellProvider(t *testing.T) {
	testConfig := config.NewCloudTestConfig()

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	defer utils.ClearFolder(tmpDir, false)
	require.Nil(t, err)

	testConfig.ConfigRoot = tmpDir
	createProvider(testConfig, "a_provider")
	createProvider(testConfig, "b_provider")

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:        "simple",
		Timeout:     15,
		PackageRoot: "./sample",
	})

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:    "simple_tagged",
		Timeout: 15,
		Source: config.ExecutionSource{
			Tags: []string{"basic"},
		},
		PackageRoot: "./sample",
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Equal(t, "there is failed tests 4", err.Error())

	require.NotNil(t, report)

	rootSuite := report.Suites[0]

	require.Len(t, rootSuite.Suites, 2)

	require.Equal(t, 2, rootSuite.Suites[0].Failures)
	require.Equal(t, 6, rootSuite.Suites[0].Tests)
	require.Len(t, rootSuite.Suites[0].Suites[0].TestCases, 3)
	require.Len(t, rootSuite.Suites[0].Suites[1].TestCases, 3)

	require.Equal(t, 2, rootSuite.Suites[0].Failures)
	require.Equal(t, 6, rootSuite.Suites[0].Tests)
	require.Equal(t, 3, len(rootSuite.Suites[1].Suites[0].TestCases))
	require.Equal(t, 3, len(rootSuite.Suites[1].Suites[1].TestCases))
}

func createProvider(testConfig *config.CloudTestConfig, name string) *config.ClusterProviderConfig {
	provider := &config.ClusterProviderConfig{
		Timeout:    100,
		Name:       name,
		NodeCount:  1,
		Kind:       "shell",
		RetryCount: 1,
		Instances:  2,
		Scripts: map[string]string{
			"config":  "echo ./.tests/config",
			"start":   "echo started",
			"prepare": "echo prepared",
			"install": "echo installed",
			"stop":    "echo stopped",
		},
		Enabled: true,
	}
	testConfig.Providers = append(testConfig.Providers, provider)
	return provider
}

func TestInvalidProvider(t *testing.T) {
	testConfig := config.NewCloudTestConfig()

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	defer utils.ClearFolder(tmpDir, false)
	require.Nil(t, err)

	testConfig.ConfigRoot = tmpDir
	createProvider(testConfig, "a_provider")
	delete(testConfig.Providers[0].Scripts, "start")

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:        "simple",
		Timeout:     2,
		PackageRoot: "./sample",
	})

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	logrus.Error(err.Error())
	require.Equal(t, "Failed to create cluster instance. Error invalid start script", err.Error())

	require.Nil(t, report)
}

func TestRequireEnvVars(t *testing.T) {
	testConfig := config.NewCloudTestConfig()

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	defer utils.ClearFolder(tmpDir, false)
	require.Nil(t, err)

	testConfig.ConfigRoot = tmpDir

	createProvider(testConfig, "a_provider")

	testConfig.Providers[0].EnvCheck = append(testConfig.Providers[0].EnvCheck, []string{
		"KUBECONFIG", "QWE",
	}...)

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:        "simple",
		Timeout:     2,
		PackageRoot: "./sample",
	})

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	logrus.Error(err.Error())
	require.Equal(t, err.Error(), "Failed to create cluster instance. Error environment variable are not specified  Required variables: [KUBECONFIG QWE]")

	require.Nil(t, report)
}

func TestRequireEnvVars_DEPS(t *testing.T) {
	testConfig := config.NewCloudTestConfig()

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	defer utils.ClearFolder(tmpDir, false)
	require.Nil(t, err)

	testConfig.ConfigRoot = tmpDir

	createProvider(testConfig, "a_provider")

	testConfig.Providers[0].EnvCheck = append(testConfig.Providers[0].EnvCheck, "PACKET_AUTH_TOKEN")
	testConfig.Providers[0].EnvCheck = append(testConfig.Providers[0].EnvCheck, "PACKET_PROJECT_ID")

	_ = os.Setenv("PACKET_AUTH_TOKEN", "token")
	_ = os.Setenv("PACKET_PROJECT_ID", "id")

	testConfig.Providers[0].Env = append(testConfig.Providers[0].Env, []string{
		"CLUSTER_RULES_PREFIX=packet",
		"CLUSTER_NAME=$(cluster-name)-$(uuid)",
		"KUBECONFIG=$(tempdir)/config",
		"TERRAFORM_ROOT=$(tempdir)/terraform",
		"TF_VAR_auth_token=${PACKET_AUTH_TOKEN}",
		"TF_VAR_master_hostname=devci-${CLUSTER_NAME}-master",
		"TF_VAR_worker1_hostname=ci-${CLUSTER_NAME}-worker1",
		"TF_VAR_project_id=${PACKET_PROJECT_ID}",
		"TF_VAR_public_key=${TERRAFORM_ROOT}/sshkey.pub",
		"TF_VAR_public_key_name=key-${CLUSTER_NAME}",
		"TF_LOG=DEBUG",
	}...)

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:        "simple",
		Timeout:     2,
		PackageRoot: "./sample",
	})

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Contains(t, err.Error(), "there is failed tests 2")

	require.NotNil(t, report)
	// Do assertions
}

func TestShellProviderShellTest(t *testing.T) {
	testConfig := config.NewCloudTestConfig()

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	defer utils.ClearFolder(tmpDir, false)
	require.Nil(t, err)

	testConfig.ConfigRoot = tmpDir
	createProvider(testConfig, "a_provider")
	createProvider(testConfig, "b_provider")

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:        "simple",
		Timeout:     15,
		PackageRoot: "./sample",
	})

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:    "simple_shell",
		Timeout: 150000,
		Kind:    "shell",
		Run: strings.Join([]string{
			"pwd",
			"ls -la",
			"echo $KUBECONFIG",
		}, "\n"),
	})

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:    "simple_shell_fail",
		Timeout: 15,
		Kind:    "shell",
		Run: strings.Join([]string{
			"pwd",
			"ls -la",
			"exit 1",
		}, "\n"),
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Equal(t, "there is failed tests 4", err.Error())

	require.NotNil(t, report)

	rootSuite := report.Suites[0]

	require.Len(t, rootSuite.Suites, 3)

	for _, executionSuite := range rootSuite.Suites {
		switch executionSuite.Name {
		case "simple":
			require.Equal(t, 2, executionSuite.Failures)
			require.Equal(t, 6, executionSuite.Tests)
			require.Len(t, executionSuite.Suites[0].TestCases, 3)
			require.Len(t, executionSuite.Suites[0].TestCases, 3)
		case "simple_shell":
			require.Equal(t, 0, executionSuite.Failures)
			require.Equal(t, 2, executionSuite.Tests)
			require.Len(t, executionSuite.Suites[0].TestCases, 1)
			require.Len(t, executionSuite.Suites[0].TestCases, 1)
		case "simple_shell_fail":
			require.Equal(t, 2, executionSuite.Failures)
			require.Equal(t, 2, executionSuite.Tests)
			require.Len(t, executionSuite.Suites[0].TestCases, 1)
			require.Len(t, executionSuite.Suites[0].TestCases, 1)
		}
	}
}

func TestUnusedClusterShutdownByMonitor(t *testing.T) {
	logKeeper := utils.NewLogKeeper()
	defer logKeeper.Stop()
	testConfig := config.NewCloudTestConfig()

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	defer utils.ClearFolder(tmpDir, false)
	require.Nil(t, err)

	testConfig.ConfigRoot = tmpDir
	createProvider(testConfig, "a_provider")
	p2 := createProvider(testConfig, "b_provider")
	p2.TestDelay = 7

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:            "simple",
		Timeout:         15,
		PackageRoot:     "./sample",
		ClusterSelector: []string{"a_provider"},
	})

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:    "simple2",
		Timeout: 15,
		Source: config.ExecutionSource{
			Tags: []string{"basic"},
		},
		PackageRoot:     "./sample",
		ClusterSelector: []string{"b_provider"},
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Contains(t, err.Error(), "there is failed tests 2")

	require.NotNil(t, report)

	rootSuite := report.Suites[0]

	require.Len(t, rootSuite.Suites, 2)

	require.Equal(t, 1, rootSuite.Suites[0].Failures)
	require.Equal(t, 3, rootSuite.Suites[0].Tests)
	require.Len(t, rootSuite.Suites[0].Suites[0].TestCases, 3)

	require.Equal(t, 1, rootSuite.Suites[1].Failures)
	require.Equal(t, 3, rootSuite.Suites[1].Tests)
	require.Equal(t, 3, len(rootSuite.Suites[1].Suites[0].TestCases))

	logKeeper.CheckMessagesOrder(t, []string{
		"All tasks for cluster group a_provider are complete. Starting cluster shutdown",
		"Destroying cluster  a_provider-",
		"Finished test execution",
	})
}

func TestMultiClusterTest(t *testing.T) {
	testConfig := config.NewCloudTestConfig()

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	defer utils.ClearFolder(tmpDir, false)
	require.Nil(t, err)

	testConfig.ConfigRoot = tmpDir
	p1 := createProvider(testConfig, "a_provider")
	p2 := createProvider(testConfig, "b_provider")
	p3 := createProvider(testConfig, "c_provider")
	p4 := createProvider(testConfig, "d_provider")

	p1.Instances = 1
	p2.Instances = 1
	p3.Instances = 1
	p4.Instances = 1

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:            "simple",
		Timeout:         15,
		PackageRoot:     "./sample",
		ClusterSelector: []string{"a_provider"},
	})

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:    "simple2",
		Timeout: 15,
		Source: config.ExecutionSource{
			Tags: []string{"interdomain"},
		},
		PackageRoot:     "./sample",
		ClusterCount:    2,
		ClusterEnv:      []string{"CFG1", "CFG2"},
		ClusterSelector: []string{"a_provider", "b_provider"},
	})
	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:    "simple3",
		Timeout: 15,
		Source: config.ExecutionSource{
			Tags: []string{"interdomain"},
		},
		PackageRoot:     "./sample",
		ClusterCount:    2,
		ClusterEnv:      []string{"CFG1", "CFG2"},
		ClusterSelector: []string{"c_provider", "d_provider"},
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Contains(t, err.Error(), "there is failed tests 3")

	require.NotNil(t, report)

	rootSuite := report.Suites[0]

	require.Len(t, rootSuite.Suites, 3)

	require.Equal(t, 1, rootSuite.Suites[0].Failures)
	require.Equal(t, 3, rootSuite.Suites[0].Tests)

	require.Equal(t, 1, rootSuite.Suites[1].Failures)
	require.Equal(t, 3, rootSuite.Suites[1].Tests)

	require.Equal(t, 1, rootSuite.Suites[2].Failures)
	require.Equal(t, 3, rootSuite.Suites[2].Tests)
}

func TestGlobalTimeout(t *testing.T) {
	testConfig := config.NewCloudTestConfig()
	testConfig.Timeout = 3

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	defer utils.ClearFolder(tmpDir, false)
	require.Nil(t, err)

	testConfig.ConfigRoot = tmpDir
	createProvider(testConfig, "a_provider")

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:        "simple",
		Timeout:     15,
		PackageRoot: "./sample",
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Equal(t, "global timeout elapsed: 3 seconds", err.Error())

	require.NotNil(t, report)

	rootSuite := report.Suites[0]

	require.Len(t, rootSuite.Suites, 1)
	require.Equal(t, 1, rootSuite.Suites[0].Failures)
	require.Equal(t, 3, rootSuite.Suites[0].Tests)
	require.Len(t, rootSuite.Suites[0].Suites[0].TestCases, 3)
}
