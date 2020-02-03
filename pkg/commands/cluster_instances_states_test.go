package commands

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"io/ioutil"

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
			clusters: []string {
				"a_provider",
				"b_provider",
			},
		},
	}
	ctx.cloudTestConfig.Timeout = 2
	ctx.cloudTestConfig.Statistics.Enabled = false

	ctx.createClusters()

	g.Expect(len(ctx.clusters)).To(BeEquivalentTo(2))
	g.Expect(len(ctx.clusters[0].instances)).To(BeEquivalentTo(1))
	ctx.startCluster(ctx.clusters[0].instances[0])
	ctx.startCluster(ctx.clusters[1].instances[0])

	<-time.After(100 * time.Millisecond)
	g.Expect(ctx.clusters[0].instances[0].state).To(BeEquivalentTo(clusterReady))
	g.Expect(ctx.clusters[1].instances[0].state).To(BeEquivalentTo(clusterCrashed))

	ctx.clusters[0].instances[0].state = clusterStarting
	ctx.destroyCluster(ctx.clusters[0].instances[0], false, false)
	g.Expect(ctx.clusters[0].instances[0].state).To(BeEquivalentTo(clusterStarting))

	ctx.clusters[0].instances[0].state = clusterStopping
	ctx.destroyCluster(ctx.clusters[0].instances[0], false, false)
	g.Expect(ctx.clusters[0].instances[0].state).To(BeEquivalentTo(clusterCrashed))
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