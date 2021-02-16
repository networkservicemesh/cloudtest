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

package parse

import (
	"strings"
	"time"

	"github.com/pkg/errors"
)

// TestEvent is a go test JSON event
type TestEvent struct {
	Time    time.Time `json:"Time"` // encodes as an RFC3339-format string
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Elapsed float64   `json:"Elapsed"` // seconds
	Output  string    `json:"Output"`
}

// TestName returns TestEvent test name without suite name
func (e *TestEvent) TestName() string {
	if split := strings.Split(e.Test, "/"); len(split) > 1 {
		return split[1]
	}
	return ""
}

// Process calls corresponding TestEventProcessor Process* method
func (e *TestEvent) Process(processor TestEventProcessor) error {
	switch e.Action {
	case "run":
		return processor.ProcessRunEvent(e)
	case "pass":
		return processor.ProcessPassEvent(e)
	case "fail":
		return processor.ProcessFailEvent(e)
	case "bench", "output":
		return processor.ProcessOutputEvent(e)
	case "skip":
		return processor.ProcessSkipEvent(e)
	default:
		return errors.Errorf("unsupported TestEvent action type: %s", e.Action)
	}
}
