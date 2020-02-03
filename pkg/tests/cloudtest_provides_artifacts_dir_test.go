// Copyright (c) 2020 Cisco and/or its affiliates.
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
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"

	"github.com/networkservicemesh/cloudtest/pkg/commands"
	"github.com/networkservicemesh/cloudtest/pkg/config"
)

func TestCloudtestProvidesArtifactsDirForEachTest(t *testing.T) {
	g := gomega.NewWithT(t)

	testConfig := &config.CloudTestConfig{}

	testConfig.Timeout = 300
	err := os.Mkdir(t.Name(), os.ModePerm)
	g.Expect(err).Should(gomega.BeNil())
	defer func() {
		_ = os.RemoveAll(t.Name())
	}()
	relativePath, err := ioutil.TempDir(t.Name(), "tmp")
	g.Expect(err).Should(gomega.BeNil())

	testConfig.ConfigRoot = relativePath
	createProvider(testConfig, "provider")
	testConfig.Providers[0].Instances = 1
	testConfig.Executions = []*config.Execution{{
		Name:        "simple",
		Timeout:     2,
		PackageRoot: "./sample",
		Source: config.ExecutionSource{
			Tags: []string{"artifacts"},
		},
	}}

	testConfig.Reporting.JUnitReportFile = JunitReport

	_, err = commands.PerformTesting(testConfig, &TestValidationFactory{}, &commands.Arguments{})
	g.Expect(err).Should(gomega.BeNil())
	content, err := ioutil.ReadFile(filepath.Join(relativePath, testConfig.Providers[0].Name+"-1", "TestArtifacts", "artifact1.txt"))
	g.Expect(err).Should(gomega.BeNil())
	g.Expect(string(content)).Should(gomega.Equal("test result"))
}
