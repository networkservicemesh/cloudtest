// Copyright (c) 2019-2020 Cisco Systems, Inc.
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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/cloudtest/pkg/commands"
	"github.com/networkservicemesh/cloudtest/pkg/config"
)

func testConfig(failedTestLimit int, source *config.ExecutionSource) *config.CloudTestConfig {
	testConfig := &config.CloudTestConfig{}
	testConfig.Timeout = 300
	testConfig.FailedTestsLimit = failedTestLimit
	createProvider(testConfig, "provider")
	testConfig.Providers[0].Instances = 1
	testConfig.Executions = []*config.Execution{{
		Name:        "simple",
		Timeout:     2,
		PackageRoot: "./sample",
		Source:      *source,
	}}
	testConfig.Reporting.JUnitReportFile = JunitReport
	return testConfig
}

func TestTerminateTestingWhenLimitReached(t *testing.T) {
	failedTestLimit := 3
	testConfig := testConfig(failedTestLimit, &config.ExecutionSource{
		Tags: []string{"failed", "passed"},
	})
	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Error(t, err)
	require.Equal(t, fmt.Sprintf("Allowed limit for failed tests is reached: %d", failedTestLimit), err.Error())
	require.NotNil(t, report)
	require.Equal(t, failedTestLimit, report.Suites[0].Failures)
}

func TestTerminateTestingWhenLimitReachedFailedOnly(t *testing.T) {
	failedTestLimit := 3
	testConfig := testConfig(failedTestLimit, &config.ExecutionSource{
		Tags: []string{"failed"},
	})
	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.Error(t, err)
	require.Equal(t, fmt.Sprintf("Allowed limit for failed tests is reached: %d", failedTestLimit), err.Error())
	require.NotNil(t, report)
	require.Equal(t, failedTestLimit, report.Suites[0].Failures)
}

func TestPassedTestsNotAffected(t *testing.T) {
	failedTestLimit := 2
	testConfig := testConfig(failedTestLimit, &config.ExecutionSource{
		Tags: []string{"passed"},
	})
	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	require.NoError(t, err)
	require.NotNil(t, report)
	require.Equal(t, 0, report.Suites[0].Failures)
}
