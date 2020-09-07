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

package suites_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/cloudtest/pkg/suites"
)

func TestFind(t *testing.T) {
	foundSuites, err := suites.Find("./samples")
	require.NoError(t, err)
	require.NotNil(t, foundSuites)
	require.Len(t, foundSuites, 6)
	sort.Slice(foundSuites, func(i, j int) bool {
		return strings.Compare(foundSuites[i].Name, foundSuites[j].Name) == -1
	})
	for i, s := range foundSuites {
		require.Equal(t, s.Name, fmt.Sprintf("TestEntryPoint%v", i+1))
	}
	require.Len(t, foundSuites[5].Tests, 4)
}
