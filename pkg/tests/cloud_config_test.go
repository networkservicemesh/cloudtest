// Copyright (c) 2019 Cisco Systems, Inc and/or its affiliates.
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
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/networkservicemesh/cloudtest/pkg/commands"
	"github.com/networkservicemesh/cloudtest/pkg/config"
)

func TestClusterConfiguration(t *testing.T) {
	g := NewWithT(t)

	var testConfig config.CloudTestConfig

	file1, err := ioutil.ReadFile("./config1.yaml")
	g.Expect(err).To(BeNil())

	err = yaml.Unmarshal(file1, &testConfig)
	g.Expect(err).To(BeNil())
	g.Expect(len(testConfig.Providers)).To(Equal(3))
	g.Expect(testConfig.Reporting.JUnitReportFile).To(Equal("./.tests/junit.xml"))
}

func TestClusterHealthCheckConfig(t *testing.T) {
	g := NewWithT(t)

	var testConfig config.CloudTestConfig

	file1, err := ioutil.ReadFile("./config1.yaml")
	g.Expect(err).To(BeNil())

	err = yaml.Unmarshal(file1, &testConfig)
	g.Expect(err).To(BeNil())
	g.Expect(len(testConfig.Providers)).To(Equal(3))
	g.Expect(testConfig.Reporting.JUnitReportFile).To(Equal("./.tests/junit.xml"))

	errChan := make(chan error, len(testConfig.HealthCheck))
	commands.RunHealthChecks(testConfig.HealthCheck, errChan)

	select {
	case err = <-errChan:
		g.Expect(err.Error()).To(ContainSubstring("Health check failed"))
	case <-time.After(5 * time.Second):
		g.Expect(false).To(BeTrue())
	}
}
