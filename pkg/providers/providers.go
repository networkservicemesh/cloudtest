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

package providers

import (
	"context"
	"time"

	"github.com/networkservicemesh/cloudtest/pkg/config"
	"github.com/networkservicemesh/cloudtest/pkg/execmanager"
	"github.com/networkservicemesh/cloudtest/pkg/k8s"
)

// InstanceOptions - instance operation parameters
type InstanceOptions struct {
	NoInstall        bool
	NoPrepare        bool
	NoMaskParameters bool
	NoStop           bool
}

// ClusterInstance - Instanceof of one cluster
// Some of cluster cloud be alive by default, it could bare metal cluster,
// and we do not need to perform any startup, shutdown code on them.
type ClusterInstance interface {
	// Return cluster Kubernetes configuration file .config location.
	GetClusterConfig() (string, error)

	// Perform startup of cluster
	Start(timeout time.Duration) (string, error)
	// Destroy cluster
	// Should destroy cluster with timeout passed, if time is left should report about error.
	Destroy(timeout time.Duration) error

	// Return root folder to store test artifacts associated with this cluster
	GetRoot() string

	// Is cluster is running right now
	IsRunning() bool
	CheckIsAlive() error
	GetID() string
}

// ClusterProvider - provides operations with clusters
type ClusterProvider interface {
	// CreateCluster - Create a cluster based on parameters
	// CreateCluster - Creates a cluster instance and put Kubernetes config file into clusterConfigRoot
	// could fully use clusterConfigRoot folder for any temporary files related to cluster.
	CreateCluster(config *config.ClusterProviderConfig, factory k8s.ValidationFactory,
		manager execmanager.ExecutionManager,
		instanceOptions InstanceOptions) (ClusterInstance, error)

	// ValidateConfig - Check if config are valid and all parameters required by this cluster are fit.
	ValidateConfig(config *config.ClusterProviderConfig) error

	// CleanupClusters - Cleaning up leaked clusters (running scripts from cluster configuration file)
	CleanupClusters(ctx context.Context, config *config.ClusterProviderConfig,
		manager execmanager.ExecutionManager, instanceOptions InstanceOptions)
}

// ClusterProviderFunction - function type to create cluster provider
type ClusterProviderFunction func(root string) ClusterProvider
