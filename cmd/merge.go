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
	// scanCmd represents the scan command.
	var scanCmd = &cobra.Command{
		Use:   "merge",
		Short: "Merge one directory to another",
		Long: `Merge directory for Exif Date Info. 

	exifsort scan [<options>...] <dir> `,
		Args: cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			src, _ := cmd.Flags().GetString("input")
			dst, _ := cmd.Flags().GetString("output")
			methodArg, _ := cmd.Flags().GetString("method")

			method, err := exifsort.ParseMethod(methodArg)
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

	var scanFlags = []cmdStringFlag{
		{"i", "input", true, "Input Directory to scan media."},
		{"m", "method", true,
			"Method to index media in output directory. <year|month|day>"},
		{"o", "output", true,
			"Output Directory to transfer media. (Must not exist.)"},
	}

	scanCmd.Flags().BoolP("quiet", "q", false,
		"Suppress line by line printing.")
	scanCmd.Flags().BoolP("summarize", "s", false,
		"Print a summary of stats when done.")

	setStringFlags(scanCmd, scanFlags)

	return scanCmd
}
