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

package controller

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/clock"
	core "k8s.io/client-go/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/heptio/ark/pkg/apis/ark/v1"
	"github.com/heptio/ark/pkg/backup"
	"github.com/heptio/ark/pkg/cloudprovider"
	"github.com/heptio/ark/pkg/generated/clientset/versioned/fake"
	informers "github.com/heptio/ark/pkg/generated/informers/externalversions"
	"github.com/heptio/ark/pkg/metrics"
	"github.com/heptio/ark/pkg/restore"
	"github.com/heptio/ark/pkg/util/collections"
	arktest "github.com/heptio/ark/pkg/util/test"
)

type fakeBackupper struct {
	mock.Mock
}

func (b *fakeBackupper) Backup(backup *v1.Backup, data, log io.Writer, actions []backup.ItemAction) error {
	args := b.Called(backup, data, log, actions)
	return args.Error(0)
}

func TestProcessBackup(t *testing.T) {
	tests := []struct {
		name             string
		key              string
		expectError      bool
		expectedIncludes []string
		expectedExcludes []string
		backup           *arktest.TestBackup
		expectBackup     bool
		allowSnapshots   bool
	}{
		{
			name:        "bad key",
			key:         "bad/key/here",
			expectError: true,
		},
		{
			name:        "lister failed",
			key:         "heptio-ark/backup1",
			expectError: true,
		},
		{
			name:         "do not process phase FailedValidation",
			key:          "heptio-ark/backup1",
			backup:       arktest.NewTestBackup().WithName("backup1").WithPhase(v1.BackupPhaseFailedValidation),
			expectBackup: false,
		},
		{
			name:         "do not process phase InProgress",
			key:          "heptio-ark/backup1",
			backup:       arktest.NewTestBackup().WithName("backup1").WithPhase(v1.BackupPhaseInProgress),
			expectBackup: false,
		},
		{
			name:         "do not process phase Completed",
			key:          "heptio-ark/backup1",
			backup:       arktest.NewTestBackup().WithName("backup1").WithPhase(v1.BackupPhaseCompleted),
			expectBackup: false,
		},
		{
			name:         "do not process phase Failed",
			key:          "heptio-ark/backup1",
			backup:       arktest.NewTestBackup().WithName("backup1").WithPhase(v1.BackupPhaseFailed),
			expectBackup: false,
		},
		{
			name:         "do not process phase other",
			key:          "heptio-ark/backup1",
			backup:       arktest.NewTestBackup().WithName("backup1").WithPhase("arg"),
			expectBackup: false,
		},
		{
			name:         "invalid included/excluded resources fails validation",
			key:          "heptio-ark/backup1",
			backup:       arktest.NewTestBackup().WithName("backup1").WithPhase(v1.BackupPhaseNew).WithIncludedResources("foo").WithExcludedResources("foo"),
			expectBackup: false,
		},
		{
			name:         "invalid included/excluded namespaces fails validation",
			key:          "heptio-ark/backup1",
			backup:       arktest.NewTestBackup().WithName("backup1").WithPhase(v1.BackupPhaseNew).WithIncludedNamespaces("foo").WithExcludedNamespaces("foo"),
			expectBackup: false,
		},
		{
			name:             "make sure specified included and excluded resources are honored",
			key:              "heptio-ark/backup1",
			backup:           arktest.NewTestBackup().WithName("backup1").WithPhase(v1.BackupPhaseNew).WithIncludedResources("i", "j").WithExcludedResources("k", "l"),
			expectedIncludes: []string{"i", "j"},
			expectedExcludes: []string{"k", "l"},
			expectBackup:     true,
		},
		{
			name:         "if includednamespaces are specified, don't default to *",
			key:          "heptio-ark/backup1",
			backup:       arktest.NewTestBackup().WithName("backup1").WithPhase(v1.BackupPhaseNew).WithIncludedNamespaces("ns-1"),
			expectBackup: true,
		},
		{
			name:         "ttl",
			key:          "heptio-ark/backup1",
			backup:       arktest.NewTestBackup().WithName("backup1").WithPhase(v1.BackupPhaseNew).WithTTL(10 * time.Minute),
			expectBackup: true,
		},
		{
			name:         "backup with SnapshotVolumes when allowSnapshots=false fails validation",
			key:          "heptio-ark/backup1",
			backup:       arktest.NewTestBackup().WithName("backup1").WithPhase(v1.BackupPhaseNew).WithSnapshotVolumes(true),
			expectBackup: false,
		},
		{
			name:           "backup with SnapshotVolumes when allowSnapshots=true gets executed",
			key:            "heptio-ark/backup1",
			backup:         arktest.NewTestBackup().WithName("backup1").WithPhase(v1.BackupPhaseNew).WithSnapshotVolumes(true),
			allowSnapshots: true,
			expectBackup:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				client          = fake.NewSimpleClientset()
				backupper       = &fakeBackupper{}
				cloudBackups    = &arktest.BackupService{}
				sharedInformers = informers.NewSharedInformerFactory(client, 0)
				logger          = arktest.NewLogger()
				pluginManager   = &MockManager{}
				clockTime, _    = time.Parse("Mon Jan 2 15:04:05 2006", "Mon Jan 2 15:04:05 2006")
			)

			c := NewBackupController(
				sharedInformers.Ark().V1().Backups(),
				client.ArkV1(),
				backupper,
				cloudBackups,
				"bucket",
				test.allowSnapshots,
				logger,
				pluginManager,
				NewBackupTracker(),
				metrics.NewServerMetrics(),
			).(*backupController)

			c.clock = clock.NewFakeClock(clockTime)

			var expiration, startTime time.Time

			if test.backup != nil {
				// add directly to the informer's store so the lister can function and so we don't have to
				// start the shared informers.
				sharedInformers.Ark().V1().Backups().Informer().GetStore().Add(test.backup.Backup)

				startTime = c.clock.Now()

				if test.backup.Spec.TTL.Duration > 0 {
					expiration = c.clock.Now().Add(test.backup.Spec.TTL.Duration)
				}

				// set up a Backup object to represent what we expect to be passed to backupper.Backup()
				backup := test.backup.DeepCopy()
				backup.Spec.IncludedResources = test.expectedIncludes
				backup.Spec.ExcludedResources = test.expectedExcludes
				backup.Spec.IncludedNamespaces = test.backup.Spec.IncludedNamespaces
				backup.Spec.SnapshotVolumes = test.backup.Spec.SnapshotVolumes
				backup.Status.Phase = v1.BackupPhaseInProgress
				backup.Status.Expiration.Time = expiration
				backup.Status.StartTimestamp.Time = startTime
				backup.Status.Version = 1
				backupper.On("Backup", backup, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				cloudBackups.On("UploadBackup", "bucket", backup.Name, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				pluginManager.On("GetBackupItemActions", backup.Name).Return(nil, nil)
				pluginManager.On("CloseBackupItemActions", backup.Name).Return(nil)
			}

			// this is necessary so the Patch() call returns the appropriate object
			client.PrependReactor("patch", "backups", func(action core.Action) (bool, runtime.Object, error) {
				if test.backup == nil {
					return true, nil, nil
				}

				patch := action.(core.PatchAction).GetPatch()
				patchMap := make(map[string]interface{})

				if err := json.Unmarshal(patch, &patchMap); err != nil {
					t.Logf("error unmarshalling patch: %s\n", err)
					return false, nil, err
				}

				phase, err := collections.GetString(patchMap, "status.phase")
				if err != nil {
					t.Logf("error getting status.phase: %s\n", err)
					return false, nil, err
				}

				res := test.backup.DeepCopy()

				// these are the fields that we expect to be set by
				// the controller
				res.Status.Version = 1
				res.Status.Expiration.Time = expiration
				res.Status.Phase = v1.BackupPhase(phase)

				// If there's an error, it's mostly likely that the key wasn't found
				// which is fine since not all patches will have them.
				completionString, err := collections.GetString(patchMap, "status.completionTimestamp")
				if err == nil {
					completionTime, err := time.Parse(time.RFC3339Nano, completionString)
					require.NoError(t, err, "unexpected completionTimestamp parsing error %v", err)
					res.Status.CompletionTimestamp.Time = completionTime
				}
				startString, err := collections.GetString(patchMap, "status.startTimestamp")
				if err == nil {
					startTime, err := time.Parse(time.RFC3339Nano, startString)
					require.NoError(t, err, "unexpected startTimestamp parsing error %v", err)
					res.Status.StartTimestamp.Time = startTime
				}

				return true, res, nil
			})

			// method under test
			err := c.processBackup(test.key)

			if test.expectError {
				require.Error(t, err, "processBackup should error")
				return
			}
			require.NoError(t, err, "processBackup unexpected error: %v", err)

			if !test.expectBackup {
				assert.Empty(t, backupper.Calls)
				assert.Empty(t, cloudBackups.Calls)
				return
			}

			actions := client.Actions()
			require.Equal(t, 2, len(actions))

			// structs and func for decoding patch content
			type StatusPatch struct {
				Expiration          time.Time      `json:"expiration"`
				Version             int            `json:"version"`
				Phase               v1.BackupPhase `json:"phase"`
				StartTimestamp      metav1.Time    `json:"startTimestamp"`
				CompletionTimestamp metav1.Time    `json:"completionTimestamp"`
			}

			type Patch struct {
				Status StatusPatch `json:"status"`
			}

			decode := func(decoder *json.Decoder) (interface{}, error) {
				actual := new(Patch)
				err := decoder.Decode(actual)

				return *actual, err
			}

			// validate Patch call 1 (setting version, expiration, and phase)
			expected := Patch{
				Status: StatusPatch{
					Version:    1,
					Phase:      v1.BackupPhaseInProgress,
					Expiration: expiration,
				},
			}

			arktest.ValidatePatch(t, actions[0], expected, decode)

			// validate Patch call 2 (setting phase, startTimestamp, completionTimestamp)
			expected = Patch{
				Status: StatusPatch{
					Phase:               v1.BackupPhaseCompleted,
					StartTimestamp:      metav1.Time{Time: c.clock.Now()},
					CompletionTimestamp: metav1.Time{Time: c.clock.Now()},
				},
			}
			arktest.ValidatePatch(t, actions[1], expected, decode)
		})
	}
}

// MockManager is an autogenerated mock type for the Manager type
type MockManager struct {
	mock.Mock
}

// CloseBackupItemActions provides a mock function with given fields: backupName
func (_m *MockManager) CloseBackupItemActions(backupName string) error {
	ret := _m.Called(backupName)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(backupName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetBackupItemActions provides a mock function with given fields: backupName, logger, level
func (_m *MockManager) GetBackupItemActions(backupName string) ([]backup.ItemAction, error) {
	ret := _m.Called(backupName)

	var r0 []backup.ItemAction
	if rf, ok := ret.Get(0).(func(string) []backup.ItemAction); ok {
		r0 = rf(backupName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]backup.ItemAction)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(backupName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CloseRestoreItemActions provides a mock function with given fields: restoreName
func (_m *MockManager) CloseRestoreItemActions(restoreName string) error {
	ret := _m.Called(restoreName)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(restoreName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetRestoreItemActions provides a mock function with given fields: restoreName, logger, level
func (_m *MockManager) GetRestoreItemActions(restoreName string) ([]restore.ItemAction, error) {
	ret := _m.Called(restoreName)

	var r0 []restore.ItemAction
	if rf, ok := ret.Get(0).(func(string) []restore.ItemAction); ok {
		r0 = rf(restoreName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]restore.ItemAction)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(restoreName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBlockStore provides a mock function with given fields: name
func (_m *MockManager) GetBlockStore(name string) (cloudprovider.BlockStore, error) {
	ret := _m.Called(name)

	var r0 cloudprovider.BlockStore
	if rf, ok := ret.Get(0).(func(string) cloudprovider.BlockStore); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cloudprovider.BlockStore)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetObjectStore provides a mock function with given fields: name
func (_m *MockManager) GetObjectStore(name string) (cloudprovider.ObjectStore, error) {
	ret := _m.Called(name)

	var r0 cloudprovider.ObjectStore
	if rf, ok := ret.Get(0).(func(string) cloudprovider.ObjectStore); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cloudprovider.ObjectStore)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CleanupClients provides a mock function
func (_m *MockManager) CleanupClients() {
	_ = _m.Called()
	return
}
