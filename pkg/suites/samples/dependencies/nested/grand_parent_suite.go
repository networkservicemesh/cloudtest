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

// Package nested comment
package nested

import "github.com/stretchr/testify/suite"

// GrandParentSuite comment
type GrandParentSuite struct {
	suite.Suite
}

// Test3 comment
func (s *GrandParentSuite) Test3() {
}

// Test4 comment
func (s *GrandParentSuite) Test4() {
}

// Test5 comment
func (s *GrandParentSuite) Test5() {
}
