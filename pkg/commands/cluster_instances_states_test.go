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

package commands

import (
	"os"
	"testing"
	"time"

	"io/ioutil"

	. "github.com/onsi/gomega"

	"github.com/networkservicemesh/cloudtest/pkg/config"
	"github.com/networkservicemesh/cloudtest/pkg/execmanager"
	"github.com/networkservicemesh/cloudtest/pkg/tests"
	"github.com/networkservicemesh/cloudtest/pkg/utils"
)

func TestClusterInstanceStates(t *testing.T) {
	g := NewWithT(t)

	tmpDir, err := ioutil.TempDir(os.TempDir(), t.Name())
	g.Expect(err).To(BeNil())
	defer utils.ClearFolder(tmpDir, false)

	testConfig := config.NewCloudTestConfig()
	testConfig.Timeout = 300
	testConfig.ConfigRoot = tmpDir
	testConfig.Providers = []*config.ClusterProviderConfig{
		createProvider(testConfig, "a_provider", "echo starting"),
		createProvider(testConfig, "b_provider", "echo starting\nexit 2"),
	}
	testConfig.Executions = append(testConfig.Executions, &config.Execution{
		Name:        "simple",
		Timeout:     15,
		PackageRoot: "./sample",
		TestsFound:  1,
	})

	ctx := executionContext{
		cloudTestConfig:  testConfig,
		manager:          execmanager.NewExecutionManager(tmpDir),
		running:          make(map[string]*testTask),
		operationChannel: make(chan operationEvent, 1),
		factory:          &tests.TestValidationFactory{},
		arguments: &Arguments{
			clusters: []string{
				"a_provider",
				"b_provider",
			},
		},
	}
	ctx.cloudTestConfig.Timeout = 2
	ctx.cloudTestConfig.Statistics.Enabled = false

	err = ctx.createClusters()
	g.Expect(err).To(BeNil())

	g.Expect(len(ctx.clusters)).To(BeEquivalentTo(2))
	g.Expect(len(ctx.clusters[0].instances)).To(BeEquivalentTo(1))
	ctx.startCluster(ctx.clusters[0].instances[0])
	ctx.startCluster(ctx.clusters[1].instances[0])

	<-time.After(100 * time.Millisecond)

	ctx.Lock()
	g.Expect(ctx.clusters[0].instances[0].state).To(BeEquivalentTo(clusterReady))
	g.Expect(ctx.clusters[1].instances[0].state).To(BeEquivalentTo(clusterCrashed))

	ctx.clusters[0].instances[0].state = clusterStarting
	ctx.Unlock()

	err = ctx.destroyCluster(ctx.clusters[0].instances[0], false, false)
	g.Expect(err).To(BeNil())

	ctx.Lock()
	g.Expect(ctx.clusters[0].instances[0].state).To(BeEquivalentTo(clusterStarting))

	ctx.clusters[0].instances[0].state = clusterStopping
	ctx.Unlock()

	err = ctx.destroyCluster(ctx.clusters[0].instances[0], false, false)
	g.Expect(err).To(BeNil())
	ctx.Lock()
	g.Expect(ctx.clusters[0].instances[0].state).To(BeEquivalentTo(clusterCrashed))
	ctx.Unlock()
}

func createProvider(testConfig *config.CloudTestConfig, name, startScript string) *config.ClusterProviderConfig {
	provider := &config.ClusterProviderConfig{
		Timeout:    100,
		Name:       name,
		NodeCount:  1,
		Kind:       "shell",
		RetryCount: 1,
		Instances:  1,
		Scripts: map[string]string{
			"config":  "echo ./.tests/config",
			"start":   startScript,
			"prepare": "echo prepared",
			"install": "echo installed",
			"stop":    "echo stopped",
		},
		Enabled: true,
	}
	testConfig.Providers = append(testConfig.Providers, provider)
	return provider
}
