/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"github.com/matchstick/exifSort/lib"
	"github.com/spf13/cobra"
	"os"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan directory for Exif Dates",
	Long: `Scan directory for Exif Date Info. 

	exifSort scan [<options>...] <directory>

	exifSort will recursively check every file in an input directory and
        then print it's exifData to stdout if possible.

        OPTIONS

	-q, --quiet
	Suppress line by line time printing

	-s,  --summary
	when done scanning print a sumamry of stats 

`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		quiet, _ := cmd.Flags().GetBool("quiet")
		summarize, _ := cmd.Flags().GetBool("summarize")

		dirPath := args[0]
		info, err := os.Stat(dirPath)
		if err != nil || info.IsDir() == false {
			fmt.Printf("Error with directory arg: %s\n", err.Error())
			return
		}
		exifSort.ScanDir(dirPath, summarize, !quiet)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().BoolP("quiet", "q", false,
		"Don't print output while scanning")
	scanCmd.Flags().BoolP("summarize", "s", false,
		"Print a summary when done scanning")
}
