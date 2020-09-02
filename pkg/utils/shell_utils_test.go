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

package utils

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProcessOutputShouldNotLostOutput(t *testing.T) {
	const expected = "output..."
	start := time.Now()
	for time.Since(start) < time.Second {
		output, err := RunCommand(context.Background(), fmt.Sprintf("echo \"%v\"", expected), "", func(s string) {}, bufio.NewWriter(&strings.Builder{}), nil, nil, true)
		require.Nil(t, err)
		require.Equal(t, expected, strings.TrimSpace(output))
	}
}
