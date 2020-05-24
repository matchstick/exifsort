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

func scanSummary(w *exifsort.WalkState) {
	fmt.Printf("Scanned Valid: %d\n", w.Valid())
	fmt.Printf("Scanned Invalid: %d\n", w.Invalid())
	fmt.Printf("Scanned Skipped: %d\n", w.Skipped())
	fmt.Printf("Scanned Total: %d\n", w.Total())

	if w.Invalid() == 0 {
		fmt.Println("No Files caused Errors")
		return
	}

	fmt.Println("Error Files were:")

	for path, msg := range w.WalkErrs() {
		fmt.Printf("\t%s\n", w.ErrStr(path, msg))
	}
}

const numScanArgs = 1

func newScanCmd() *cobra.Command {
	// scanCmd represents the scan command.
	var scanCmd = &cobra.Command{
		Use:   "scan",
		Short: "Scan directory for Exif Dates",
		Long: `Scan directory for Exif Date Info. 

	exifsort scan [<options>...] <dir>

	exifsort will recursively check every file in an input directory and
        then print it's exifData to stdout if possible.

	ARGUMENTS

	src
	Input directory of media files`,
		Args: cobra.MinimumNArgs(numScanArgs),
		Run: func(cmd *cobra.Command, args []string) {
			quiet, _ := cmd.Flags().GetBool("quiet")
			summarize, _ := cmd.Flags().GetBool("summarize")

			dirPath := args[0]
			info, err := os.Stat(dirPath)
			if err != nil || !info.IsDir() {
				fmt.Printf("Error with directory arg: %s\n", err.Error())
				return
			}
			w := exifsort.ScanDir(dirPath, !quiet)
			if summarize {
				scanSummary(&w)
			}
		},
	}

	scanCmd.Flags().BoolP("quiet", "q", false,
		"Suppress line by line printing.")
	scanCmd.Flags().BoolP("summarize", "s", false,
		"Print a summary of stats when done.")

	return scanCmd
}
