// Copyright 2020 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"fmt"
	"io/ioutil"

	"go.uber.org/zap"

	"github.com/kapetaniosci/pipe/pkg/app/piped/analysisprovider/metrics/datadog"
	"github.com/kapetaniosci/pipe/pkg/app/piped/analysisprovider/metrics/prometheus"
	"github.com/kapetaniosci/pipe/pkg/config"
)

type Factory struct {
	logger *zap.Logger
}

func NewFactory(logger *zap.Logger) *Factory {
	return &Factory{logger: logger}
}

// NewProvider generates an appropriate provider according to analysis provider config.
func (f *Factory) NewProvider(providerCfg *config.AnalysisProvider) (provider Provider, err error) {
	switch {
	case providerCfg.Prometheus != nil:
		cfg := providerCfg.Prometheus
		// TODO: Decide the way to authenticate.
		/*		username, err := ioutil.ReadFile(cfg.UsernameFile)
				if err != nil {
					return nil, err
				}
				password, err := ioutil.ReadFile(cfg.PasswordFile)
				if err != nil {
					return nil, err
				}
				provider, err = prometheus.NewProvider(cfg.Address, string(username), string(password))
		*/
		provider, err = prometheus.NewProvider(cfg.Address, "", "", f.logger)
		if err != nil {
			return
		}
	case providerCfg.Datadog != nil:
		cfg := providerCfg.Datadog
		apiKey, err := ioutil.ReadFile(cfg.APIKeyFile)
		if err != nil {
			return nil, err
		}
		applicationKey, err := ioutil.ReadFile(cfg.ApplicationKeyFile)
		if err != nil {
			return nil, err
		}
		provider, err = datadog.NewProvider(cfg.Address, string(apiKey), string(applicationKey))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("any of providers config not found")
	}
	return provider, nil
}
