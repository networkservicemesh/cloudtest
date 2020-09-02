// +build interdomain

package sample

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestInterDomainPass(t *testing.T) {
	logrus.Infof("Passed test")
}

func TestInterdomainCheck(t *testing.T) {
	g := NewWithT(t)

	require.NotEmpty(t, os.Getenv("CFG1"))
}
func TestInterdomainFail(t *testing.T) {
	logrus.Infof("Failed test")

	t.FailNow()
}
