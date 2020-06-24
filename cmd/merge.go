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
	"fmt"
	"os"

	exifsort "github.com/matchstick/exifsort/lib"
	"github.com/spf13/cobra"
)

func mergeSummary(m *exifsort.Merger) {
	fmt.Printf("## Merged files: %d\n", len(m.Merged))

	if len(m.Removed) != 0 {
		fmt.Printf("## Duplicates Removed %d:\n", len(m.Removed))

		for _, path := range m.Removed {
			fmt.Printf("##\t%s\n", path)
		}
	}

	if len(m.Errors) != 0 {
		fmt.Printf("## Errors were %d:\n", len(m.Errors))

		for path, err := range m.Errors {
			fmt.Printf("##\t%s: (%s)\n", path, err)
		}
	}
}

func mergeExecute(src string, dst string, methodArg string, actionArg string,
	matchStr string) {
	method, err := exifsort.MethodParse(methodArg)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	action, err := exifsort.ActionParse(actionArg)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	merger := exifsort.NewMerger(src, dst, action, method, matchStr)

	err = merger.Merge(os.Stdout)
	if err != nil {
		fmt.Printf("Merge Error: %s\n", err.Error())
		return
	}

	mergeSummary(merger)
}

func newMergeCmd() *cobra.Command {
	const minMergeArgs = 4
	// scanCmd represents the scan command.
	var mergeCmd = &cobra.Command{
		Use:   "merge",
		Short: "Merge one sorted directory to another sorted directory",
		Long: `Merge one sorted directory to another sorted directory.

	exifsort merge <src> <dir> 

	src
	directory or json file to receive media to sort

	dst
	directory to create to transfer media
`,
		Args: cobra.MinimumNArgs(minMergeArgs),
		Run: func(cmd *cobra.Command, args []string) {
			src := args[0]
			dst := args[1]
			actionArg := args[2]
			methodArg := args[3]

			mergeExecute(src, dst, methodArg, actionArg, "")
		},
	}

	return mergeCmd
}
