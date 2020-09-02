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

package commands

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/cloudtest/pkg/config"
)

func TestImportAll(t *testing.T) {
	testConfig := &config.CloudTestConfig{
		Imports: []string{"samples/.*"},
	}
	err := performImport(testConfig)
	require.Nil(t, err)
	files, _ := ioutil.ReadDir(testConfig.Imports[0][:len(testConfig.Imports[0])-1])
	require.Len(t, files, len(testConfig.Executions))
}

func TestImportPattern(t *testing.T) {
	testConfig := &config.CloudTestConfig{
		Imports: []string{"samples/.*2"},
	}
	err := performImport(testConfig)
	require.Nil(t, err)
	require.Len(t, testConfig.Executions, 1)
}

func TestSpecificImport(t *testing.T) {
	testConfig := &config.CloudTestConfig{
		Imports: []string{"samples/execution1.yaml"},
	}
	err := performImport(testConfig)
	require.Nil(t, err)
	require.Len(t, testConfig.Executions, 1)
}
