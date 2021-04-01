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

// Package testentry provides a builder class to build model.TestEntry from parse.TestEvent sequence
package testentry

import (
	"fmt"
	"os"

	"github.com/networkservicemesh/cloudtest/pkg/execmanager"
	"github.com/networkservicemesh/cloudtest/pkg/model"
	"github.com/networkservicemesh/cloudtest/pkg/suites/parse"
)

// Builder is a builder class to build model.TestEntry from parse.TestEvent sequence
type Builder struct {
	manager       execmanager.ExecutionManager
	clusterTaskID string
	suiteEntry    *model.TestEntry
	testEntry     *model.TestEntry
	file          *os.File
}

// NewBuilder returns a new Builder
func NewBuilder(
	name string,
	suiteEntry *model.TestEntry,
	manager execmanager.ExecutionManager,
	clusterTaskID string,
) *Builder {
	return &Builder{
		manager:       manager,
		clusterTaskID: clusterTaskID,
		suiteEntry:    suiteEntry,
		testEntry: &model.TestEntry{
			Name:            name,
			ExecutionConfig: suiteEntry.ExecutionConfig,
			RunScript:       suiteEntry.RunScript,
			Kind:            model.GoTestKind,
		},
	}
}

// Build finishes building and returns a new model.TestEntry
func (b *Builder) Build() *model.TestEntry {
	if b.file != nil {
		_ = b.processStatusEvent(&parse.TestEvent{
			Time: b.suiteEntry.Started.Add(b.suiteEntry.Duration),
		}, model.StatusTimeout)
	}
	return b.testEntry
}

// ProcessRunEvent processes "run" parse.TestEvent
func (b *Builder) ProcessRunEvent(testEvent *parse.TestEvent) (err error) {
	if b.testEntry.Started.IsZero() {
		b.testEntry.Started = testEvent.Time
	}

	fileName := fmt.Sprintf("%s-%s", b.suiteEntry.Name, b.testEntry.Name)
	if fileName, b.file, err = b.manager.OpenFileTest(b.clusterTaskID, fileName, "run"); err != nil {
		return err
	}

	b.testEntry.Executions = append(b.testEntry.Executions, model.TestEntryExecution{
		OutputFile: fileName,
		Retry:      len(b.testEntry.Executions) + 1,
	})

	return nil
}

// ProcessPassEvent processes "pass" parse.TestEvent
func (b *Builder) ProcessPassEvent(testEvent *parse.TestEvent) error {
	return b.processStatusEvent(testEvent, model.StatusSuccess)
}

// ProcessFailEvent processes "fail" parse.TestEvent
func (b *Builder) ProcessFailEvent(testEvent *parse.TestEvent) error {
	return b.processStatusEvent(testEvent, model.StatusFailed)
}

// ProcessOutputEvent processes "output" parse.TestEvent
func (b *Builder) ProcessOutputEvent(testEvent *parse.TestEvent) error {
	if b.file != nil {
		_, err := b.file.WriteString(testEvent.Output)
		return err
	}
	return nil
}

// ProcessSkipEvent processes "skip" parse.TestEvent
func (b *Builder) ProcessSkipEvent(testEvent *parse.TestEvent) error {
	b.testEntry.SkipMessage = testEvent.Output

	return b.processStatusEvent(testEvent, model.StatusSkipped)
}

func (b *Builder) processStatusEvent(testEvent *parse.TestEvent, status model.Status) error {
	b.testEntry.Duration = testEvent.Time.Sub(b.testEntry.Started)

	b.testEntry.Executions[len(b.testEntry.Executions)-1].Status = status
	b.testEntry.Status = status

	_ = b.file.Close()
	b.file = nil

	return nil
}
