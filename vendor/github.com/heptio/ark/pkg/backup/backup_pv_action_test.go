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

package backup

import (
	"testing"

	"github.com/heptio/ark/pkg/apis/ark/v1"
	"github.com/heptio/ark/pkg/kuberesource"
	arktest "github.com/heptio/ark/pkg/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestBackupPVAction(t *testing.T) {
	pvc := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{},
		},
	}

	backup := &v1.Backup{}

	a := NewBackupPVAction(arktest.NewLogger())

	// no spec.volumeName should result in no error
	// and no additional items
	_, additional, err := a.Execute(pvc, backup)
	assert.NoError(t, err)
	assert.Len(t, additional, 0)

	// empty spec.volumeName should result in no error
	// and no additional items
	pvc.Object["spec"].(map[string]interface{})["volumeName"] = ""
	_, additional, err = a.Execute(pvc, backup)
	assert.NoError(t, err)
	assert.Len(t, additional, 0)

	// non-empty spec.volumeName should result in
	// no error and an additional item for the PV
	pvc.Object["spec"].(map[string]interface{})["volumeName"] = "myVolume"
	_, additional, err = a.Execute(pvc, backup)
	require.NoError(t, err)
	require.Len(t, additional, 1)
	assert.Equal(t, ResourceIdentifier{GroupResource: kuberesource.PersistentVolumes, Name: "myVolume"}, additional[0])
}
