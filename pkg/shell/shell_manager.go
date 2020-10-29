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

package shell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/edwarnicke/exechelper"
	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/cloudtest/pkg/config"
	"github.com/networkservicemesh/cloudtest/pkg/execmanager"
	"github.com/networkservicemesh/cloudtest/pkg/providers"
	"github.com/networkservicemesh/cloudtest/pkg/utils"
)

// Manager - allow to perform shell command executions with variable and parameter substitutions.
type Manager interface {
	// GetConfigLocation - detect if KUBECONFIG variable is passed and return its value.
	GetConfigLocation() string
	// RunCmd - execute a command, operation with extra env
	RunCmd(context context.Context, operation string, script []string, env []string) (string, error)
	// RunRead - execute a command, operation with extra env and read response into variable
	RunRead(context context.Context, operation string, script []string, env []string) (string, error)
	// PrintEnv - print environment variables into string
	PrintEnv(processedEnv []string) string
	// PrintArgs - print arguments to string
	PrintArgs() string
	EnvironmentManager
}

type shellInterface struct {
	id      string
	config  *config.ClusterProviderConfig
	manager execmanager.ExecutionManager
	params  providers.InstanceOptions
	environmentManager
}

// NewManager - creates a new shell manager
func NewManager(manager execmanager.ExecutionManager, id string, config *config.ClusterProviderConfig,
	params providers.InstanceOptions) Manager {
	return &shellInterface{
		manager: manager,
		id:      id,
		config:  config,
		params:  params,
	}
}

// RunCmd -  command in context and add appropriate execution output file.
func (si *shellInterface) RunCmd(context context.Context, operation string, script, env []string) (string, error) {
	fileName, _, err := si.runCmd(context, operation, script, env, false)
	return fileName, err
}

// Run command in context and add appropriate execution output file.
func (si *shellInterface) RunRead(context context.Context, operation string, script, env []string) (string, error) {
	_, response, err := si.runCmd(context, operation, script, env, true)
	return response, err
}

type bufferedWriter struct {
	buffer   io.Writer
	original io.Writer
}

func (b *bufferedWriter) Write(p []byte) (n int, err error) {
	_, _ = b.buffer.Write(p)
	return b.original.Write(p)
}

var _ io.Writer = (*bufferedWriter)(nil)

func (si *shellInterface) runCmd(context context.Context, operation string, script, env []string, returnResult bool) (string, string, error) {
	fileName, fileRef, err := si.manager.OpenFile(si.id, operation)
	if err != nil {
		logrus.Errorf("failed to %s system for testing of cluster %s %v", operation, si.config.Name, err)
		return fileName, "", err
	}

	defer func() { _ = fileRef.Close() }()

	writer := bufio.NewWriter(fileRef)

	finalOut := ""
	for _, cmd := range script {
		if strings.TrimSpace(cmd) == "" {
			continue
		}

		cmdEnv := append(si.processedEnv, env...)
		printableEnv := si.PrintEnv(env)

		_, _ = writer.WriteString(fmt.Sprintf("%s: %v\nENV={\n%v\n}\n", operation, cmd, printableEnv))
		_ = writer.Flush()

		logrus.Infof("%s: %s => %s", operation, si.id, cmd)
		finalEnv := append(os.Environ(), cmdEnv...)
		environment := map[string]string{}
		for _, k := range finalEnv {
			key, value, err := utils.ParseVariable(k)
			if err != nil {
				return "", "", err
			}
			environment[key] = value
		}
		finalCmd, err := utils.SubstituteVariable(cmd, environment, si.finalArgs)
		if err != nil {
			return "", "", err
		}

		stdOut := new(strings.Builder)

		err = exechelper.Run(finalCmd,
			exechelper.WithContext(context),
			exechelper.WithStderr(&bufferedWriter{buffer: stdOut, original: writer}),
			exechelper.WithStdout(writer),
			exechelper.WithEnvirons(finalEnv...),
			exechelper.WithDir(""))

		if err != nil {
			_, _ = writer.WriteString(fmt.Sprintf("error running command: %v\n", err))
			_ = writer.Flush()
			return fileName, "", err
		}
		if returnResult {
			finalOut += stdOut.String()
		}
	}
	return fileName, finalOut, nil
}
func (si *shellInterface) PrintEnv(processedEnv []string) string {
	printableEnv := strings.Builder{}
	for _, cmdEnvValue := range processedEnv {
		varName, varValue, _ := utils.ParseVariable(cmdEnvValue)

		if !si.params.NoMaskParameters {
			// We need to check if value contains or not some of check env variables and replace their values for safity
			for _, ce := range si.config.EnvCheck {
				envValue := os.Getenv(ce)
				varValue = strings.Replace(varValue, envValue, "****", -1)
			}
		}
		_, _ = printableEnv.WriteString(fmt.Sprintf("%s=%s\n", varName, varValue))
	}
	return printableEnv.String()
}
func (si *shellInterface) PrintArgs() string {
	printableEnv := strings.Builder{}

	_, _ = printableEnv.WriteString("Arguments:\n")

	for varName, varValue := range si.finalArgs {
		if !si.params.NoMaskParameters {
			// We need to check if value contains or not some of check env variables and replace their values for safity
			for _, ce := range si.config.EnvCheck {
				envValue := os.Getenv(ce)
				varValue = strings.Replace(varValue, envValue, "****", -1)
			}
		}
		_, _ = printableEnv.WriteString(fmt.Sprintf("%s=%s\n", varName, varValue))
	}
	return printableEnv.String()
}
