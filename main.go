package main

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // This is required for GKE authentication to work properly

	"github.com/networkservicemesh/cloudtest/pkg/commands"
)

func main() {
	logrus.SetFormatter(&nested.Formatter{})
	logrus.SetLevel(logrus.TraceLevel)

	commands.ExecuteCloudTest()
}
