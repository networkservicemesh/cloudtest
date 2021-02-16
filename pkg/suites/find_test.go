// Copyright (c) 2021 Doc.ai and/or its affiliates.
//
// Copyright (c) 2019-2021 Cisco Systems, Inc.
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

package suites_test

import (
	"io/ioutil"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/networkservicemesh/cloudtest/pkg/suites"
)

func testResults() map[string][]string {
	bytes, err := ioutil.ReadFile("./samples/test_result.yaml")
	if err != nil {
		return nil
	}

	results := make(map[string][]string)
	if err := yaml.Unmarshal(bytes, &results); err != nil {
		return nil
	}

	return results
}

func TestFind(t *testing.T) {
	foundSuites, err := suites.Find("./samples")
	require.NoError(t, err)

	foundSuitesMap := make(map[string][]string)
	for _, suite := range foundSuites {
		sort.Strings(suite.Tests)
		foundSuitesMap[suite.Name] = suite.Tests
	}

	require.Equal(t, testResults(), foundSuitesMap)
}
