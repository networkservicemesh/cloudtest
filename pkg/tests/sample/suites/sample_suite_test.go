package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SuiteExample struct {
	suite.Suite
}

func (s *SuiteExample) SetupSuite() {
	println("SETUP")
}

func (s *SuiteExample) TearDownSuite() {
	println("TEARDOWN")
}

func (s *SuiteExample) Test1() {
}

func (s *SuiteExample) Test2() {
}

func (s *SuiteExample) Test3() {
}

func (s *SuiteExample) Test4() {
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(SuiteExample))
}
