/*
Copyright © 2020 Michael Rubin <mhr@neverthere.org>

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

package cmd

import (
	"github.com/spf13/cobra"
)

func newFilterCmd() *cobra.Command {
	const minFilterArgs = 4
	// scanCmd represents the scan command.
	var filterCmd = &cobra.Command{
		Use:   "filter",
		Short: "Transfer the contents that match regex in one sorted directory to another sorted directory",
		Long: `Transfer the contents that match a regex in one sorted directory to another sorted directory.

	exifsort filter <src> <dir> <regexp>

	src
	directory or json file to receive media to sort

	dst
	directory to create to transfer media

	regexp
	regular expression to match filename to merge to dst
`,
		Args: cobra.MinimumNArgs(minFilterArgs),
		Run: func(cmd *cobra.Command, args []string) {
			src := args[0]
			dst := args[1]
			methodArg := args[2]
			matchStr := args[3]

			mergeExecute(src, dst, methodArg, matchStr)
		},
	}

	return filterCmd
}
