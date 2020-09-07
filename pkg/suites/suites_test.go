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
	suites, err := suites.Find("./samples")
	require.NoError(t, err)
	require.NotNil(t, suites)
	require.Len(t, suites, 6)
	sort.Slice(suites, func(i, j int) bool {
		return strings.Compare(suites[i].Name, suites[j].Name) == -1
	})
	for i, s := range suites {
		require.Equal(t, s.Name, fmt.Sprintf("TestEntryPoint%v", i+1))
	}
	require.Len(t, suites[5].Tests, 4)
}
