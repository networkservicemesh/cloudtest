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
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

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
	g := NewWithT(t)
	failedTestLimit := 3
	testConfig := testConfig(failedTestLimit, &config.ExecutionSource{
		Tags: []string{"failed", "passed"},
	})
	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	g.Expect(err).ShouldNot(BeNil())
	g.Expect(err.Error()).To(Equal(fmt.Sprintf("failed tests limit is reached: %d", failedTestLimit)))
	g.Expect(report).ShouldNot(BeNil())
	g.Expect(report.Suites[0].Failures).To(Equal(failedTestLimit))
}

func TestTerminateTestingWhenLimitReachedFailedOnly(t *testing.T) {
	g := NewWithT(t)
	failedTestLimit := 3
	testConfig := testConfig(failedTestLimit, &config.ExecutionSource{
		Tags: []string{"failed"},
	})
	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	g.Expect(err).ShouldNot(BeNil())
	g.Expect(err.Error()).To(Equal(fmt.Sprintf("failed tests limit is reached: %d", failedTestLimit)))
	g.Expect(report).ShouldNot(BeNil())
	g.Expect(report.Suites[0].Failures).To(Equal(failedTestLimit))
}

func TestPassedTestsNotAffected(t *testing.T) {
	g := NewWithT(t)
	failedTestLimit := 2
	testConfig := testConfig(failedTestLimit, &config.ExecutionSource{
		Tags: []string{"passed"},
	})
	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	g.Expect(err).Should(BeNil())
	g.Expect(report).ShouldNot(BeNil())
	g.Expect(report.Suites[0].Failures).To(Equal(0))
}

func TestNumberOfTestsIsLessThanLimit(t *testing.T) {
	g := NewWithT(t)
	failedTestLimit := 11
	testConfig := testConfig(failedTestLimit, &config.ExecutionSource{
		Tags: []string{"failed", "passed"},
	})
	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	g.Expect(err).ShouldNot(BeNil())
	g.Expect(err.Error()).To(Equal(fmt.Sprintf("number of tests is less than the failed tests limit")))
	g.Expect(report).Should(BeNil())
}
