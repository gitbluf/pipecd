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

package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"

	"github.com/pipe-cd/pipe/pkg/app/piped/toolregistry"
	"github.com/pipe-cd/pipe/pkg/config"
)

const (
	LabelManagedBy          = "pipecd.dev/managed-by"           // Always be piped.
	LabelApplication        = "pipecd.dev/application"          // The application this resource belongs to.
	LabelVariant            = "pipecd.dev/variant"              // Variant name: primary, stage, baseline
	LabelCommitHash         = "pipecd.dev/commit-hash"          // Hash value of the deployed commit.
	LabelResourceKey        = "pipecd.dev/resource-key"         // The resource key generated by apiVersion, namespace and name. e.g. apps/v1/Deployment/namespace/demo-app
	LabelOriginalAPIVersion = "pipecd.dev/original-api-version" // The api version defined in git configuration. e.g. apps/v1
	ManagedByPiped          = "piped"

	kustomizationFileName = "kustomization.yaml"
)

type TemplatingMethod string

const (
	TemplatingMethodHelm      TemplatingMethod = "helm"
	TemplatingMethodKustomize TemplatingMethod = "kustomize"
	TemplatingMethodNone      TemplatingMethod = "none"
)

type Provider interface {
	ManifestLoader
	Applier
}

type ManifestLoader interface {
	// LoadManifests renders and loads all manifests for application.
	LoadManifests(ctx context.Context) ([]Manifest, error)
}

type Applier interface {
	// Apply does applying application manifests by using the tool specified in Input.
	Apply(ctx context.Context) error
	// ApplyManifests does applying the given manifests.
	ApplyManifests(ctx context.Context, manifests []Manifest) error
	// Delete deletes the given resource from Kubernetes cluster.
	Delete(ctx context.Context, key ResourceKey) error
}

type provider struct {
	appDir  string
	repoDir string
	input   config.KubernetesDeploymentInput
	logger  *zap.Logger

	kubectl          *Kubectl
	kustomize        *Kustomize
	helm             *Helm
	templatingMethod TemplatingMethod
	initOnce         sync.Once
	initErr          error
}

func NewProvider(appDir, repoDir string, input config.KubernetesDeploymentInput, logger *zap.Logger) Provider {
	return &provider{
		appDir:  appDir,
		repoDir: repoDir,
		input:   input,
		logger:  logger.Named("kubernetes-provider"),
	}
}

func NewManifestLoader(appDir, repoDir string, input config.KubernetesDeploymentInput, logger *zap.Logger) ManifestLoader {
	return NewProvider(appDir, repoDir, input, logger)
}

func (p *provider) init(ctx context.Context) {
	p.templatingMethod = determineTemplatingMethod(p.input, p.appDir)

	switch p.templatingMethod {
	case TemplatingMethodHelm:
		p.helm, p.initErr = p.findHelm(ctx, p.input.HelmVersion)

	case TemplatingMethodKustomize:
		p.kustomize, p.initErr = p.findKustomize(ctx, p.input.KustomizeVersion)

	case TemplatingMethodNone:
		p.kubectl, p.initErr = p.findKubectl(ctx, p.input.KubectlVersion)

	default:
		p.initErr = fmt.Errorf("unsupport templating method %v", p.templatingMethod)
	}
}

func (p *provider) LoadManifests(ctx context.Context) (manifests []Manifest, err error) {
	p.initOnce.Do(func() { p.init(ctx) })
	if p.initErr != nil {
		return nil, p.initErr
	}

	switch p.templatingMethod {
	case TemplatingMethodHelm:
		return nil, nil

	case TemplatingMethodKustomize:
		return nil, nil

	case TemplatingMethodNone:
		manifests, err = LoadPlainYAMLMannifests(ctx, p.appDir, p.input.Manifests)

	default:
		err = fmt.Errorf("unsupport templating method %v", p.templatingMethod)
	}
	return
}

func (p *provider) Apply(ctx context.Context) error {
	return nil
}

func (p *provider) ApplyManifests(ctx context.Context, manifests []Manifest) (err error) {
	p.initOnce.Do(func() { p.init(ctx) })
	if p.initErr != nil {
		return p.initErr
	}

	switch p.templatingMethod {
	case TemplatingMethodHelm:
		return nil

	case TemplatingMethodKustomize:
		return nil

	case TemplatingMethodNone:
		err = p.kubectl.Apply(ctx, manifests)

	default:
		err = fmt.Errorf("unsupport templating method %v", p.templatingMethod)
	}
	return
}

func (p *provider) Delete(ctx context.Context, k ResourceKey) (err error) {
	p.initOnce.Do(func() { p.init(ctx) })
	if p.initErr != nil {
		return p.initErr
	}

	switch p.templatingMethod {
	case TemplatingMethodHelm:
		return nil

	case TemplatingMethodKustomize:
		return nil

	case TemplatingMethodNone:
		err = p.kubectl.Delete(ctx, k)

	default:
		err = fmt.Errorf("unsupport templating method %v", p.templatingMethod)
	}
	return
}

func (p *provider) findKubectl(ctx context.Context, version string) (*Kubectl, error) {
	path, installed, err := toolregistry.DefaultRegistry().Kubectl(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("no kubectl %s (%v)", version, err)
	}
	if installed {
		p.logger.Info(fmt.Sprintf("kubectl %s has just been installed because of no pre-installed binary for that version", version))
	}
	return NewKubectl(path), nil
}

func (p *provider) findKustomize(ctx context.Context, version string) (*Kustomize, error) {
	path, installed, err := toolregistry.DefaultRegistry().Kustomize(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("no kustomize %s (%v)", version, err)
	}
	if installed {
		p.logger.Info(fmt.Sprintf("kustomize %s has just been installed because of no pre-installed binary for that version", version))
	}
	return NewKustomize(path), nil
}

func (p *provider) findHelm(ctx context.Context, version string) (*Helm, error) {
	path, installed, err := toolregistry.DefaultRegistry().Helm(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("no helm %s (%v)", version, err)
	}
	if installed {
		p.logger.Info(fmt.Sprintf("helm %s has just been installed because of no pre-installed binary for that version", version))
	}
	return NewHelm(path), nil
}

func determineTemplatingMethod(input config.KubernetesDeploymentInput, appDirPath string) TemplatingMethod {
	if input.HelmChart != nil {
		return TemplatingMethodHelm
	}
	if len(input.HelmValueFiles) > 0 {
		return TemplatingMethodHelm
	}
	if input.HelmVersion != "" {
		return TemplatingMethodHelm
	}
	if _, err := os.Stat(filepath.Join(appDirPath, kustomizationFileName)); err == nil {
		return TemplatingMethodKustomize
	}
	return TemplatingMethodNone
}
