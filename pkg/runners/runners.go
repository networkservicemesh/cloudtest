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

package runners

import (
	"bufio"
	"context"
)

// TestRunner - describes a way to execute tests.
type TestRunner interface {
	// Run - run tests with timeout context, environment
	Run(timeoutCtx context.Context, env []string, writer *bufio.Writer) error
	// GetCmdLine - return created command line, if applicable.
	GetCmdLine() string
}
