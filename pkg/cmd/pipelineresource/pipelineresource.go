// Copyright © 2019 The Tekton Authors.
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

package pipelineresource

import (
	"github.com/spf13/cobra"
	"github.com/tektoncd/cli/pkg/cli"
	"github.com/tektoncd/cli/pkg/flags"
)

func Command(p cli.Params) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resource",
		Aliases: []string{"res", "resources"},
		Short:   "Manage pipeline resources",
		Annotations: map[string]string{
			"commandType": "main",
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return flags.InitParams(p, cmd)
		},
	}

	flags.AddTektonOptions(cmd)
	cmd.AddCommand(
		createCommand(p),
		deleteCommand(p),
		describeCommand(p),
		listCommand(p),
	)
	cmd.Deprecated = "PipelineResource commands are deprecated, they will be removed soon as it get removed from API."

	return cmd
}
