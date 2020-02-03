package tests

import (
	"context"

	"github.com/networkservicemesh/cloudtest/pkg/config"
	"github.com/networkservicemesh/cloudtest/pkg/k8s"
)

type TestValidationFactory struct {
}

type testValidator struct {
	location string
	config   *config.ClusterProviderConfig
}

func (v *testValidator) WaitValid(context context.Context) error {
	return nil
}

func (v *testValidator) Validate() error {
	// Validation is passed for now
	return nil
}

func (*TestValidationFactory) CreateValidator(config *config.ClusterProviderConfig, location string) (k8s.KubernetesValidator, error) {
	return &testValidator{
		config:   config,
		location: location,
	}, nil
}
