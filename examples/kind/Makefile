# Copyright 2019 VMware, Inc.
# SPDX-License-Identifier: Apache-2.0
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at:
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ifeq ($(KIND_CLUSTER_NAME),)
	KIND_CLUSTER_NAME := "kind-demo"
endif

ifeq ($(CONFIG_LOCATION),)
	CONFIG_LOCATION := "./.results/config"
endif

.PHONY: kind-install
kind-install:
	GO111MODULE="on" go get -u sigs.k8s.io/kind@master

.PHONY: kind-start
kind-start:
	@kind get clusters | grep $(KIND_CLUSTER_NAME)  >/dev/null 2>&1 && exit 0 || \
		( kind create cluster --name="$(KIND_CLUSTER_NAME)" --config ./kind.yaml --wait 120s && \
		until \
			kubectl taint node $(KIND_CLUSTER_NAME)-control-plane node-role.kubernetes.io/master:NoSchedule- ; \
		do echo "Waiting for the cluster to come up" && sleep 3; done )

.PHONY: kind-export-kubeconfig
kind-export-kubeconfig:
	@touch $(CONFIG_LOCATION); \
	kind get kubeconfig --name $(KIND_CLUSTER_NAME) > $(CONFIG_LOCATION); \
	echo "$(CONFIG_LOCATION)"


.PHONY: kind-stop
kind-stop:
	@kind delete cluster --name="$(KIND_CLUSTER_NAME)"

.PHONY: kind-restart
kind-restart: kind-stop kind-start
	@echo "kind cluster restarted"
