// Copyright (c) 2019 Cisco Systems, Inc and/or its affiliates.
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

package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/cloudtest/pkg/utils"
)

func TestVariableSubstitutions(t *testing.T) {
	env := map[string]string{
		"KUBECONFIG": "~/.kube/config",
	}

	args := map[string]string{
		"cluster-name":  "idd",
		"provider-name": "name",
		"random":        "r1",
		"uuid":          "uu-uu",
		"tempdir":       "/tmp",
		"zone-selector": "zone",
	}

	var1, err := utils.SubstituteVariable("qwe ${KUBECONFIG} $(uuid) BBB", env, args)
	require.Nil(t, err)
	require.Equal(t, "qwe ~/.kube/config uu-uu BBB", var1)
}

func TestParseCommandLine1(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		require.Equal(t, []string{"a", "b", "c"}, utils.ParseCommandLine("a b c"))
	})

	t.Run("spaces", func(t *testing.T) {
		require.Equal(t, []string{"a b", "c"}, utils.ParseCommandLine("a\\ b c"))
	})

	t.Run("strings", func(t *testing.T) {
		require.Equal(t, []string{"a", "b    ", "c"}, utils.ParseCommandLine("a \"b    \" c"))
	})

	t.Run("empty_arg", func(t *testing.T) {
		require.Equal(t, []string{"a", "-N", ""}, utils.ParseCommandLine("a 	-N \"\""))
	})
}

func TestParseCommandLine2(t *testing.T) {
	cmdLine := utils.ParseCommandLine("go test ./test/integration -test.timeout 300s -count 1 --run \"^(TestNSMHealLocalDieNSMD)$\" --tags \"basic recover usecase\" --test.v")
	require.Len(t, cmdLine, 12)
}

func TestParseCommandLine3(t *testing.T) {
	cmdLine := utils.ParseCommandLine("go test ./test/integration -test.timeout 300s -count 1 --run \"^(TestNSMHealLocalDieNSMD)$\\\\z\" --tags \"basic recover usecase\" --test.v")
	require.Len(t, cmdLine, 12)
	require.Equal(t, cmdLine[8], "^(TestNSMHealLocalDieNSMD)$\\z")
}
