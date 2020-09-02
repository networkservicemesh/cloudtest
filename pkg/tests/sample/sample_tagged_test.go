// +build basic

package sample

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestPassTag(t *testing.T) {
	logrus.Infof("Passed test")
}

func TestFailTag(t *testing.T) {
	logrus.Infof("Failed test")
	t.FailNow()
}

func TestTimeoutTag(t *testing.T) {
	logrus.Infof("test timeout for 5 seconds")
	<-time.After(5 * time.Second)
}
