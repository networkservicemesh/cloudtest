// +build interdomain

package sample

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sirupsen/logrus"
)

func TestInterDomainPass(t *testing.T) {
	logrus.Infof("Passed test")
}

func TestInterdomainCheck(t *testing.T) {
	require.NotEmpty(t, os.Getenv("CFG1"))
	require.NotEmpty(t, os.Getenv("CFG2"))
}
func TestInterdomainFail(t *testing.T) {
	logrus.Infof("Failed test")

	t.FailNow()
}
