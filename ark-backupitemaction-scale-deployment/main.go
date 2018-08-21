/*
Copyright 2017 the Heptio Ark contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/heptio/ark/pkg/apis/ark/v1"
	"github.com/heptio/ark/pkg/backup"
	"github.com/heptio/ark/pkg/plugin"
)

func main() {
	impl := &ScaleDeploymentsToZeroReplicas{
		log: plugin.NewLogger(),
	}

	plugin.Serve(plugin.NewBackupItemActionPlugin(impl))
}

// ScaleDeploymentsToZeroReplicas is a backup item action plugin for Heptio Ark.
type ScaleDeploymentsToZeroReplicas struct {
	log logrus.FieldLogger
}

// AppliesTo returns a backup.ResourceSelector that applies to deployments only.
func (p *ScaleDeploymentsToZeroReplicas) AppliesTo() (backup.ResourceSelector, error) {
	return backup.ResourceSelector{
		IncludedResources: []string{"deployments.apps"},
	}, nil
}

// Execute sets .spec.replicas to "0".
func (p *ScaleDeploymentsToZeroReplicas) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []backup.ResourceIdentifier, error) {
	p.log.Info("Running ScaleDeploymentsToZeroReplicas backup item action")
	defer p.log.Info("Done running ScaleDeploymentsToZeroReplicas backup item action")

	if err := unstructured.SetNestedField(item.UnstructuredContent(), "0", "spec", "replicas"); err != nil {
		p.log.WithError(err).Error("Error setting .spec.replicase")
		return nil, nil, err
	}

	return item, nil, nil
}
