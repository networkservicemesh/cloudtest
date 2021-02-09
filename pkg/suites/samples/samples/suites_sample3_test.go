// Copyright (c) 2021 Doc.ai and/or its affiliates.
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

package samples

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/networkservicemesh/cloudtest/pkg/suites/samples/dependencies"
)

func TestLibEntryPoint1(t *testing.T) {
	suite.Run(t, new(dependencies.LibSuite))
}

func TestLibEntryPoint2(t *testing.T) {
	s := new(dependencies.LibSuite)
	suite.Run(t, s)
}

func TestLibEntryPoint3(t *testing.T) {
	s := &dependencies.LibSuite{}
	suite.Run(t, s)
}

func TestLibEntryPoint4(t *testing.T) {
	suite.Run(t, &dependencies.LibSuite{})
}

