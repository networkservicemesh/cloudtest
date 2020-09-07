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

package samples

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type Suite2 struct {
	suite.Suite
}

func (s *Suite2) Test1() {
}
func (s *Suite2) Test2() {
}
func (s *Suite2) Test3() {
}
func (s *Suite2) Test4() {
}
func (s *Suite2) Setup() {
}

func FakeStart(t *testing.T) {
	suite.Run(t, &Suite2{})
}

func TestEntryPoint5(t *testing.T) {
	suite.Run(t, &Suite2{})
}

func TestEntryPoint6(t *testing.T) {
	suite.Run(t, new(Suite2))
}
