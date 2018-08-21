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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/heptio/ark/pkg/apis/ark/v1"
	pkgbackup "github.com/heptio/ark/pkg/backup"
	"github.com/heptio/ark/pkg/client"
	"github.com/heptio/ark/pkg/cmd"
	"github.com/heptio/ark/pkg/cmd/util/output"
	"github.com/heptio/ark/pkg/restic"
)

func NewDescribeCommand(f client.Factory, use string) *cobra.Command {
	var (
		listOptions   metav1.ListOptions
		volumeDetails bool
	)

	c := &cobra.Command{
		Use:   use + " [NAME1] [NAME2] [NAME...]",
		Short: "Describe backups",
		Run: func(c *cobra.Command, args []string) {
			arkClient, err := f.Client()
			cmd.CheckError(err)

			var backups *v1.BackupList
			if len(args) > 0 {
				backups = new(v1.BackupList)
				for _, name := range args {
					backup, err := arkClient.Ark().Backups(f.Namespace()).Get(name, metav1.GetOptions{})
					cmd.CheckError(err)
					backups.Items = append(backups.Items, *backup)
				}
			} else {
				backups, err = arkClient.ArkV1().Backups(f.Namespace()).List(listOptions)
				cmd.CheckError(err)
			}

			first := true
			for _, backup := range backups.Items {
				deleteRequestListOptions := pkgbackup.NewDeleteBackupRequestListOptions(backup.Name, string(backup.UID))
				deleteRequestList, err := arkClient.ArkV1().DeleteBackupRequests(f.Namespace()).List(deleteRequestListOptions)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error getting DeleteBackupRequests for backup %s: %v\n", backup.Name, err)
				}

				opts := restic.NewPodVolumeBackupListOptions(backup.Name, string(backup.UID))
				podVolumeBackupList, err := arkClient.ArkV1().PodVolumeBackups(f.Namespace()).List(opts)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error getting PodVolumeBackups for backup %s: %v\n", backup.Name, err)
				}

				s := output.DescribeBackup(&backup, deleteRequestList.Items, podVolumeBackupList.Items, volumeDetails)
				if first {
					first = false
					fmt.Print(s)
				} else {
					fmt.Printf("\n\n%s", s)
				}
			}
			cmd.CheckError(err)
		},
	}

	c.Flags().StringVarP(&listOptions.LabelSelector, "selector", "l", listOptions.LabelSelector, "only show items matching this label selector")
	c.Flags().BoolVar(&volumeDetails, "volume-details", volumeDetails, "display details of restic volume backups")

	return c
}
