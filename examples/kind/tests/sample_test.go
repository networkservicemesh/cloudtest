// Copyright (c) 2019-2020 Cisco Systems, Inc.
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

package sample

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestPass(t *testing.T) {
	logrus.Infof("Passed test:" + os.Getenv("KUBECONFIG"))
}

func TestFail(t *testing.T) {
	logrus.Infof("Failed test: " + os.Getenv("KUBECONFIG"))

	t.FailNow()
}

func TestTimeout(t *testing.T) {
	logrus.Infof("test timeout for 5 seconds:" + os.Getenv("KUBECONFIG"))
	<-time.After(5 * time.Second)
}
