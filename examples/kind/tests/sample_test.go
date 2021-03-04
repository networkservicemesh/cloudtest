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
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
}

func (s *Suite) TestPass() {
	logrus.Infof("Passed test:" + os.Getenv("KUBECONFIG"))
}

func (s *Suite) TestFail() {
	logrus.Infof("Failed test: " + os.Getenv("KUBECONFIG"))

	s.T().FailNow()
}

type FailedSuite struct {
	suite.Suite
}

func (s *FailedSuite) SetupSuite() {
	logrus.Infof("Failed suite: " + os.Getenv("KUBECONFIG"))

	s.T().FailNow()
}

func (s *FailedSuite) TestPass() {
	logrus.Infof("Passed test:" + os.Getenv("KUBECONFIG"))
}

func (s *FailedSuite) TestFail() {
	logrus.Infof("Failed test: " + os.Getenv("KUBECONFIG"))

	s.T().FailNow()
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func TestFailedSuite(t *testing.T) {
	suite.Run(t, new(FailedSuite))
}

func TestPass(t *testing.T) {
	logrus.Infof("Passed test:" + os.Getenv("KUBECONFIG"))
}

func TestFail(t *testing.T) {
	logrus.Infof("Failed test: " + os.Getenv("KUBECONFIG"))

	t.FailNow()
}

func TestTimeout(t *testing.T) {
	logrus.Infof("test timeout for 10 seconds:" + os.Getenv("KUBECONFIG"))
	<-time.After(20 * time.Second)
}
