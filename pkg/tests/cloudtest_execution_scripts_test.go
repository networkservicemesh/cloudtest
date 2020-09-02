// Copyright (c) 2019-2020 Cisco and/or its affiliates.
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
	"path"
	"testing"

	"github.com/networkservicemesh/cloudtest/pkg/commands"
	"github.com/networkservicemesh/cloudtest/pkg/config"
	"github.com/networkservicemesh/cloudtest/pkg/utils"
)

func TestAfterWorksCorrectly(t *testing.T) {
	testConfig := &config.CloudTestConfig{}

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	defer utils.ClearFolder(tmpDir, false)
	require.Nil(t, err)

	testConfig.ConfigRoot = tmpDir
	provider := &config.ClusterProviderConfig{
		Timeout:    100,
		Name:       "provider",
		NodeCount:  1,
		Kind:       "shell",
		RetryCount: 1,
		Instances:  1,
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

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:    "test1",
		Timeout: 15,
		Kind:    "shell",
		Run:     "echo first",
		Env:     []string{"A=worked", "B=$(test-name)"},
		After:   "echo ${B} ${A}",
	})
	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:    "test2",
		Timeout: 15,
		Kind:    "shell",
		Run:     "echo second",
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.NoError(t, err)
	require.NotNil(t, report)

	path := path.Join(tmpDir, provider.Name+"-1", "007-test2-run.log")
	content, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(content), "After worked")
}

func TestBeforeWorksCorrectly(t *testing.T) {
	testConfig := &config.CloudTestConfig{}

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	defer utils.ClearFolder(tmpDir, false)
	require.Nil(t, err)

	testConfig.ConfigRoot = tmpDir
	provider := &config.ClusterProviderConfig{
		Timeout:    100,
		Name:       "provider",
		NodeCount:  1,
		Kind:       "shell",
		RetryCount: 1,
		Instances:  1,
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

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:    "test1",
		Timeout: 15,
		Kind:    "shell",
		Run:     "echo first",
		Env:     []string{"A=worked", "B=$(test-name)"},
		Before:  "echo ${B} ${A}",
	})
	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:    "test2",
		Timeout: 15,
		Kind:    "shell",
		Run:     "echo second",
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.NoError(t, err)
	require.NotNil(t, report)

	path := path.Join(tmpDir, provider.Name+"-1", "006-test1-run.log")
	content, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(content), "Before worked")
}
