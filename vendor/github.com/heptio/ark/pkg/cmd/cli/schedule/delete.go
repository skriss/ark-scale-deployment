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

package schedule

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/heptio/ark/pkg/client"
	"github.com/heptio/ark/pkg/cmd"
)

func NewDeleteCommand(f client.Factory, use string) *cobra.Command {
	c := &cobra.Command{
		Use:   fmt.Sprintf("%s NAME", use),
		Short: "Delete a schedule",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			arkClient, err := f.Client()
			cmd.CheckError(err)

			name := args[0]

			err = arkClient.ArkV1().Schedules(f.Namespace()).Delete(name, nil)
			cmd.CheckError(err)

			fmt.Printf("Schedule %q deleted\n", name)
		},
	}

	return c
}
