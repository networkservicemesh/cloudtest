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
	"bufio"
	"bytes"
	"encoding/json"
	"io"
)

// Event is a TestEvent with error
type Event struct {
	Err error

	*TestEvent
}

// Events generates an Event channel from reader
func Events(reader io.Reader) <-chan *Event {
	scanner := bufio.NewScanner(reader)
	scanner.Split(splitJSON)

	ch := make(chan *Event)
	go func() {
		defer close(ch)
		for scanner.Scan() {
			var testEvent TestEvent
			switch err := json.Unmarshal(scanner.Bytes(), &testEvent); {
			case err == nil:
				ch <- &Event{TestEvent: &testEvent}
			case err != io.EOF:
				ch <- &Event{Err: err}
				fallthrough
			default:
				return
			}
		}
	}()
	return ch
}

func splitJSON(data []byte, atEOF bool) (advance int, token []byte, err error) {
	begin := bytes.IndexByte(data, '{')
	if begin < 0 {
		// We don't have token start.
		return 0, nil, nil
	}

	if advance, token, err = bufio.ScanLines(data[begin:], atEOF); token != nil {
		advance += begin
	}

	return advance, token, err
}