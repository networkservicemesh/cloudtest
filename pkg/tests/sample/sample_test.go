package sample

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestPass(t *testing.T) {
	logrus.Infof("Passed test")
}

func TestFail(t *testing.T) {

	logrus.Infof("Failed test")

	t.FailNow()
}

func TestTimeout(t *testing.T) {
	logrus.Infof("test timeout for 5 seconds")
	<-time.After(5 * time.Second)
}
