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

func newMergeCmd() *cobra.Command {
	const minMergeArgs = 3
	// scanCmd represents the scan command.
	var scanCmd = &cobra.Command{
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
			methodArg := args[2]

			method, err := exifsort.MethodParse(methodArg)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}

			err = exifsort.MergeCheck(src, method)
			if err != nil {
				fmt.Printf("Input Dir Error: %s\n", err.Error())
				return
			}

			err = exifsort.MergeCheck(dst, method)
			if err != nil {
				fmt.Printf("Output Dir Error: %s\n", err.Error())
				return
			}

			err = exifsort.Merge(src, dst, method, os.Stdout)
			if err != nil {
				fmt.Printf("Merge Error: %s\n", err.Error())
				return
			}
		},
	}

	return scanCmd
}
