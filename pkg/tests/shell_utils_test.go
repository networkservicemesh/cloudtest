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

	. "github.com/onsi/gomega"

	"github.com/networkservicemesh/cloudtest/pkg/utils"
)

func TestVariableSubstitutions(t *testing.T) {
	g := NewWithT(t)

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
	g.Expect(err).To(BeNil())
	g.Expect(var1).To(Equal("qwe ~/.kube/config uu-uu BBB"))
}

func TestParseCommandLine1(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		g := NewWithT(t)
		g.Expect(utils.ParseCommandLine("a b c")).To(Equal([]string{"a", "b", "c"}))
	})

	t.Run("spaces", func(t *testing.T) {
		g := NewWithT(t)
		g.Expect(utils.ParseCommandLine("a\\ b c")).To(Equal([]string{"a b", "c"}))
	})

	t.Run("strings", func(t *testing.T) {
		g := NewWithT(t)
		g.Expect(utils.ParseCommandLine("a \"b    \" c")).To(Equal([]string{"a", "b    ", "c"}))
	})

	t.Run("empty_arg", func(t *testing.T) {
		g := NewWithT(t)
		g.Expect(utils.ParseCommandLine("a 	-N \"\"")).To(Equal([]string{"a", "-N", ""}))
	})
}

func TestParseCommandLine2(t *testing.T) {
	g := NewWithT(t)

	cmdLine := utils.ParseCommandLine("go test ./test/integration -test.timeout 300s -count 1 --run \"^(TestNSMHealLocalDieNSMD)$\" --tags \"basic recover usecase\" --test.v")
	g.Expect(len(cmdLine)).To(Equal(12))
}
func TestParseCommandLine3(t *testing.T) {
	g := NewWithT(t)

	cmdLine := utils.ParseCommandLine("go test ./test/integration -test.timeout 300s -count 1 --run \"^(TestNSMHealLocalDieNSMD)$\\\\z\" --tags \"basic recover usecase\" --test.v")
	g.Expect(len(cmdLine)).To(Equal(12))
	g.Expect(cmdLine[8]).To(Equal("^(TestNSMHealLocalDieNSMD)$\\z"))
}
