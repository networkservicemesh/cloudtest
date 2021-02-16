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

package parse_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/cloudtest/pkg/suites/parse"
)

func TestEvents(t *testing.T) {
	file, err := os.Open("sample.log")
	require.NoError(t, err)

	var testEvents []*parse.TestEvent
	for event := range parse.Events(file) {
		require.NoError(t, event.Err)
		testEvents = append(testEvents, event.TestEvent)
	}

	require.Len(t, testEvents, 16)
}
