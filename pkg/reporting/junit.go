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

package reporting

import "encoding/xml"

const (
	// TimeCommentFormat is a format for printing readable time suite comment
	TimeCommentFormat = "Suite was running for %v"
)

// JUnitFile - JUnitFile
type JUnitFile struct {
	XMLName xml.Name `xml:"testsuites"`
	Suites  []*Suite
}

// Suite - Suite
type Suite struct {
	XMLName     xml.Name    `xml:"testsuite"`
	Tests       int         `xml:"tests,attr"`
	Failures    int         `xml:"failures,attr"`
	Time        string      `xml:"time,attr"`
	Name        string      `xml:"name,attr"`
	Properties  []*Property `xml:"properties>property,omitempty"`
	TimeComment string      `xml:",comment"`
	TestCases   []*TestCase
	Suites      []*Suite
}

// SuiteDetails holds additional information about test suite.
type SuiteDetails struct {
	XMLName       xml.Name `xml:"details"`
	FormattedTime string   `xml:"formatted_time,attr"`
}

// TestCase - TestCase
type TestCase struct {
	XMLName     xml.Name     `xml:"testcase"`
	Classname   string       `xml:"classname,attr"`
	Name        string       `xml:"name,attr"`
	Time        string       `xml:"time,attr"`
	Cluster     string       `xml:"cluster_instance,attr"`
	SkipMessage *SkipMessage `xml:"skipped,omitempty"`
	Failure     *Failure     `xml:"failure,omitempty"`
}

// SkipMessage - JUnitSkipMessage contains the reason why a testcase was skipped.
type SkipMessage struct {
	Message string `xml:"message,attr"`
}

// Property -  represents a key/value pair used to define properties.
type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// Failure -  contains data related to a failed test.
type Failure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}
