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

	"github.com/networkservicemesh/cloudtest/pkg/config"
	"github.com/networkservicemesh/cloudtest/pkg/utils"

	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/cloudtest/pkg/commands"
)

func TestOnlyRun(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "cloud-test-temp")
	require.NoError(t, err)
	defer utils.ClearFolder(tmpDir, false)
	testConfig := config.NewCloudTestConfig()
	testConfig.ConfigRoot = tmpDir
	testConfig.Timeout = 300
	createProvider(testConfig, "a_provider")
	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:        "simple",
		Timeout:     15,
		PackageRoot: "./sample",
	})
	testConfig.Reporting.JUnitReportFile = JunitReport

	testConfig.OnlyRun = []string{"TestPass"}

	report, err := commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	if err != nil {
		logrus.Errorf("Testing failed: %v", err)
	}
	require.NoError(t, err)
	require.NotNil(t, report)

	rootSuite := report.Suites[0]
	require.Len(t, rootSuite.Suites, 1)
	require.Equal(t, 0, rootSuite.Suites[0].Failures)
	require.Equal(t, 1, rootSuite.Suites[0].Tests)
}
