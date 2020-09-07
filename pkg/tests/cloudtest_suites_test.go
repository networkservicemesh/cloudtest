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

package tests

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/cloudtest/pkg/commands"
	"github.com/networkservicemesh/cloudtest/pkg/config"
	"github.com/networkservicemesh/cloudtest/pkg/utils"
)

func TestCloudtestCanWorkWithSuites(t *testing.T) {
	testConfig := config.NewCloudTestConfig()

	testConfig.Timeout = 300

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	require.NoError(t, err)
	defer utils.ClearFolder(tmpDir, false)
	testConfig.ConfigRoot = tmpDir
	createProvider(testConfig, "a_provider")
	testConfig.MinSuiteSize = 3

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:        "simple",
		Timeout:     15,
		PackageRoot: "./sample",
		Source:      config.ExecutionSource{Tags: []string{"suite"}},
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.NoError(t, err)
	require.NotNil(t, report)

	const testName = "TestRunSuite"

	provider1SuiteTestCount := getSuiteRunTestsCount(t, path.Join(tmpDir, testConfig.Providers[0].Name+"-1"), testName)
	provider2SuiteTestCount := getSuiteRunTestsCount(t, path.Join(tmpDir, testConfig.Providers[0].Name+"-2"), testName)

	if provider1SuiteTestCount != 4 && provider2SuiteTestCount != 4 {
		require.FailNow(t, "one of providers should handle 4 sub-tests")
	}

	require.Equal(t, 4, provider2SuiteTestCount+provider1SuiteTestCount)
}

func TestCloudtestCanWorkWithSuitesSplit(t *testing.T) {
	testConfig := config.NewCloudTestConfig()

	testConfig.Timeout = 300
	testConfig.MinSuiteSize = 2

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	require.NoError(t, err)
	defer utils.ClearFolder(tmpDir, false)

	testConfig.ConfigRoot = tmpDir
	createProvider(testConfig, "a_provider")

	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:        "simple",
		Timeout:     15,
		PackageRoot: "./sample",
		Source:      config.ExecutionSource{Tags: []string{"suite"}},
	})

	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.NoError(t, err)
	require.NotNil(t, report)

	const testName = "TestRunSuite"

	provider1SuiteTestCount := getSuiteRunTestsCount(t, path.Join(tmpDir, testConfig.Providers[0].Name+"-1"), testName)
	provider2SuiteTestCount := getSuiteRunTestsCount(t, path.Join(tmpDir, testConfig.Providers[0].Name+"-2"), testName)
	require.Equal(t, 2, provider1SuiteTestCount)
	require.Equal(t, 2, provider2SuiteTestCount)
}

func getSuiteRunTestsCount(t *testing.T, dir, testName string) int {
	files, err := ioutil.ReadDir(dir)
	for _, file := range files {
		require.NoError(t, err)
		if strings.Contains(file.Name(), testName) {
			bytes, err := ioutil.ReadFile(path.Join(dir, file.Name()))
			require.NoError(t, err)
			log := string(bytes)
			require.Equal(t, 1, getPatternMatchCount(log, "SETUP"))
			require.Equal(t, 1, getPatternMatchCount(log, "TEARDOWN"))
			return getPatternMatchCount(log, testName)/2 - 2
		}
	}
	return 0
}

func getPatternMatchCount(source, pattern string) int {
	regexpPattern := regexp.MustCompile(pattern)
	return len(regexpPattern.FindAllStringIndex(source, -1))
}
