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

package cloudprovider

import (
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

// SnapshotService exposes Ark-specific operations for snapshotting and restoring block
// volumes.
type SnapshotService interface {
	// CreateSnapshot triggers a snapshot for the specified cloud volume and tags it with metadata.
	// it returns the cloud snapshot ID, or an error if a problem is encountered triggering the snapshot via
	// the cloud API.
	CreateSnapshot(volumeID, volumeAZ string, tags map[string]string) (string, error)

	// CreateVolumeFromSnapshot triggers a restore operation to create a new cloud volume from the specified
	// snapshot and volume characteristics. Returns the cloud volume ID, or an error if a problem is
	// encountered triggering the restore via the cloud API.
	CreateVolumeFromSnapshot(snapshotID, volumeType, volumeAZ string, iops *int64) (string, error)

	// DeleteSnapshot triggers a deletion of the specified Ark snapshot via the cloud API. It returns an
	// error if a problem is encountered triggering the deletion via the cloud API.
	DeleteSnapshot(snapshotID string) error

	// GetVolumeInfo gets the type and IOPS (if applicable) from the cloud API.
	GetVolumeInfo(volumeID, volumeAZ string) (string, *int64, error)

	// GetVolumeID returns the cloud provider specific identifier for the PersistentVolume.
	GetVolumeID(pv runtime.Unstructured) (string, error)

	// SetVolumeID sets the cloud provider specific identifier for the PersistentVolume.
	SetVolumeID(pv runtime.Unstructured, volumeID string) (runtime.Unstructured, error)
}

const (
	volumeCreateWaitTimeout  = 30 * time.Second
	volumeCreatePollInterval = 1 * time.Second
)

type snapshotService struct {
	blockStore BlockStore
}

var _ SnapshotService = &snapshotService{}

// NewSnapshotService creates a snapshot service using the provided block store
func NewSnapshotService(blockStore BlockStore) SnapshotService {
	return &snapshotService{
		blockStore: blockStore,
	}
}

func (sr *snapshotService) CreateVolumeFromSnapshot(snapshotID string, volumeType string, volumeAZ string, iops *int64) (string, error) {
	volumeID, err := sr.blockStore.CreateVolumeFromSnapshot(snapshotID, volumeType, volumeAZ, iops)
	if err != nil {
		return "", err
	}

	// wait for volume to be ready (up to a maximum time limit)
	ticker := time.NewTicker(volumeCreatePollInterval)
	defer ticker.Stop()

	timeout := time.NewTimer(volumeCreateWaitTimeout)

	for {
		select {
		case <-timeout.C:
			return "", errors.Errorf("timeout reached waiting for volume %v to be ready", volumeID)
		case <-ticker.C:
			if ready, err := sr.blockStore.IsVolumeReady(volumeID, volumeAZ); err == nil && ready {
				return volumeID, nil
			}
		}
	}
}

func (sr *snapshotService) CreateSnapshot(volumeID, volumeAZ string, tags map[string]string) (string, error) {
	return sr.blockStore.CreateSnapshot(volumeID, volumeAZ, tags)
}

func (sr *snapshotService) DeleteSnapshot(snapshotID string) error {
	return sr.blockStore.DeleteSnapshot(snapshotID)
}

func (sr *snapshotService) GetVolumeInfo(volumeID, volumeAZ string) (string, *int64, error) {
	return sr.blockStore.GetVolumeInfo(volumeID, volumeAZ)
}

func (sr *snapshotService) GetVolumeID(pv runtime.Unstructured) (string, error) {
	return sr.blockStore.GetVolumeID(pv)
}

func (sr *snapshotService) SetVolumeID(pv runtime.Unstructured, volumeID string) (runtime.Unstructured, error) {
	return sr.blockStore.SetVolumeID(pv, volumeID)
}
