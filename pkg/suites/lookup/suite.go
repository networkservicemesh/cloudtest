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

package lookup

// Suite is a go test suite with parent Suite and tests names
type Suite struct {
	parent *Suite
	tests  []string
}

// GetTests returns set of the Suite and its parent Suite tests names
func (s *Suite) GetTests() (tests []string) {
	dupls := make(map[string]struct{})
	for _, test := range s.tests {
		tests = append(tests, test)
		dupls[test] = struct{}{}
	}
	if s.parent != nil {
		for _, test := range s.parent.GetTests() {
			if _, ok := dupls[test]; !ok {
				tests = append(tests, test)
			}
		}
	}
	return tests
}
