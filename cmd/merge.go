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

func mergeExecute(src string, dst string, action exifsort.Action, matchStr string) {
	merger := exifsort.NewMerger(src, dst, action, matchStr)

	err := merger.Merge(os.Stdout)
	if err != nil {
		fmt.Printf("Merge Error: %s\n", err.Error())
		return
	}

	mergeSummary(merger)
}

func mergeLongHelp() string {
	return `Merge one sorted directory to another sorted directory.

	exifsort merge <src> <dir> 

	src
	directory or json file to receive media to sort

	dst
	directory to create to transfer media
`
}

func newMergeActionCmd(action exifsort.Action) *cobra.Command {
	const numMethodCmdArgs = 2

	actionStr := action.String()

	actionCmd := &cobra.Command{
		Use:   actionStr,
		Short: "Merge by " + actionStr,
		// Very long help message so we moved it to a func.
		Long: mergeLongHelp(),
		Args: cobra.MinimumNArgs(numMethodCmdArgs),
		Run: func(cmd *cobra.Command, args []string) {
			src := args[0]
			dst := args[1]

			mergeExecute(src, dst, action, "")
		},
	}

	return actionCmd
}

func newMergeCmd() *cobra.Command {
	// scanCmd represents the scan command.
	rootCmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge one sorted directory to another sorted directory",
		Long:  mergeLongHelp(),
	}

	for _, action := range exifsort.Actions() {
		actionCmd := newMergeActionCmd(action)
		rootCmd.AddCommand(actionCmd)
	}

	return rootCmd
}
