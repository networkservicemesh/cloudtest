package runners

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/edwarnicke/exechelper"

	"github.com/networkservicemesh/cloudtest/pkg/model"
	"github.com/networkservicemesh/cloudtest/pkg/shell"
)

type SuiteRunner struct {
	cmd        string
	envManager shell.EnvironmentManager
	test       *model.TestEntry
}

func (s *SuiteRunner) Run(ctx context.Context, envs []string, writer *bufio.Writer) error {
	envs = append(envs, s.envManager.GetProcessedEnv()...)
	err := exechelper.Run(s.cmd,
		exechelper.WithStdout(writer),
		exechelper.WithStderr(writer),
		exechelper.WithContext(ctx),
		exechelper.WithDir(s.test.ExecutionConfig.PackageRoot))
	exechelper.WithEnvirons(envs...)
	return err
}

func (s *SuiteRunner) GetCmdLine() string {
	return s.cmd
}

var _ TestRunner = (*SuiteRunner)(nil)

func NewSuiteRunner(ids string, test *model.TestEntry, timeout time.Duration) *SuiteRunner {
	pattern := strings.Join(test.Suite.Tests, "|")
	cmdLine := fmt.Sprintf("go test . -testify.m \"%v\" -test.timeout %v -count 1 --run \"^(%s)$\\\\z\" --tags \"%s\" --test.v",
		pattern, timeout, test.Suite.Name, test.Tags)

	envMgr := shell.NewEnvironmentManager()
	_ = envMgr.ProcessEnvironment(ids, "gotest", os.TempDir(), test.ExecutionConfig.Env, map[string]string{})
	return &SuiteRunner{
		test:       test,
		cmd:        cmdLine,
		envManager: envMgr,
	}
}
