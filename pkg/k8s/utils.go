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

package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Utils - basic Kubernetes utils.
type Utils struct {
	config    *rest.Config
	clientset *kubernetes.Clientset
}

// NewK8sUtils - Creates a new k8s utils with config file.
func NewK8sUtils(configPath string) (*Utils, error) {
	utils := &Utils{}
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}

	utils.config = config
	utils.clientset, err = kubernetes.NewForConfig(utils.config)

	return utils, err
}

// GetNodes - return a list of kubernetes nodes.
func (u *Utils) GetNodes() ([]v1.Node, error) {
	nodes, err := u.clientset.CoreV1().Nodes().List(context.TODO(), v12.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodes.Items, nil
}
