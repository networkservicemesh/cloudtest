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
	"github.com/networkservicemesh/cloudtest/pkg/commands"
	"github.com/networkservicemesh/cloudtest/pkg/config"
	. "github.com/onsi/gomega"
	"testing"
)

func TestTerminateTestingAfterFailuresLimitReached(t *testing.T) {
	g := NewWithT(t)
	failedTestLimit := 2

	testConfig := &config.CloudTestConfig{}
	testConfig.Timeout = 300
	testConfig.FailedTestsLimit = failedTestLimit
	createProvider(testConfig, "provider")
	testConfig.Providers[0].Instances = 1
	testConfig.Executions = []*config.Execution{{
		Name:        "simple",
		Timeout:     2,
		PackageRoot: "./sample",
		Source: config.ExecutionSource{
			Tags: []string{"failed_limit"},
		},
	}}
	testConfig.Reporting.JUnitReportFile = JunitReport

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	g.Expect(err).ShouldNot(BeNil())
	g.Expect(err.Error()).To(Equal(fmt.Sprintf("failed tests limit is reached: %d", failedTestLimit)))
	g.Expect(report).ShouldNot(BeNil())
	g.Expect(report.Suites[0].Failures).To(Equal(failedTestLimit))
}
