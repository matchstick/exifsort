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

	exifsort "github.com/matchstick/exifsort/lib"
	"github.com/spf13/cobra"
)

func filterLongHelp() string {
	return `Transfer the contents that match a regex in one sorted directory to another sorted directory.

	exifsort filter <src> <dir> <regexp>

	src
	directory or json file to receive media to sort

	dst
	directory to create to transfer media

	regexp
	regular expression to match filename to merge to dst
`
}

func newFilterMethodCmd(action exifsort.Action,
	method exifsort.Method) *cobra.Command {
	const numMethodCmdArgs = 3

	actionStr := action.String()
	methodStr := method.String()

	return &cobra.Command{
		Use:   method.String(),
		Short: fmt.Sprintf("Transfer by %s then merge by %s", actionStr, methodStr),
		// Very long help message so we moved it to a func.
		Long: mergeLongHelp(),
		Args: cobra.MinimumNArgs(numMethodCmdArgs),
		Run: func(cmd *cobra.Command, args []string) {
			src := args[0]
			dst := args[1]
			filter := args[2]

			mergeExecute(src, dst, method, action, filter)
		},
	}
}

func newFilterActionCmd(action exifsort.Action) *cobra.Command {
	actionStr := action.String()

	actionCmd := &cobra.Command{
		Use:   actionStr,
		Short: "Filter by " + actionStr,
		// Very long help message so we moved it to a func.
		Long: filterLongHelp(),
	}

	for _, method := range exifsort.Methods() {
		methodCmd := newFilterMethodCmd(action, method)
		actionCmd.AddCommand(methodCmd)
	}

	return actionCmd
}

func newFilterCmd() *cobra.Command {
	// scanCmd represents the scan command.
	rootCmd := &cobra.Command{
		Use:   "filter",
		Short: "Transfer the contents that match regex between sorted directories",
		Long:  filterLongHelp(),
	}

	for _, action := range exifsort.Actions() {
		actionCmd := newFilterActionCmd(action)
		rootCmd.AddCommand(actionCmd)
	}

	return rootCmd
}
