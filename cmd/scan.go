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

func scanSummary(s *exifsort.Scanner) {
	fmt.Printf("Scanned Valid: %d\n", s.NumValid())
	fmt.Printf("Scanned Invalid: %d\n", s.NumInvalid())
	fmt.Printf("Scanned Skipped: %d\n", s.NumSkipped())
	fmt.Printf("Scanned Total: %d\n", s.NumTotal())

	if s.NumInvalid() == 0 {
		fmt.Println("No Files caused Errors")
		return
	}

	fmt.Println("Error Files were:")

	for path, err := range s.Errors {
		fmt.Printf("\t%s\n", exifsort.ErrStr(path, err))
	}
}

func scanSave(s *exifsort.Scanner, json string) {
	if json == "" {
		return
	}

	err := s.Save(json)
	if err != nil {
		fmt.Printf("json file %s Error:%s\n", json, err.Error())
		return
	}
}

func newScanCmd() *cobra.Command {
	// scanCmd represents the scan command.
	var scanCmd = &cobra.Command{
		Use:   "scan",
		Short: "Scan directory for Exif Dates",
		Long: `Scan directory for Exif Date Info. 

	exifsort scan [<options>...] <dir> `,
		Args: cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			quiet, _ := cmd.Flags().GetBool("quiet")
			summarize, _ := cmd.Flags().GetBool("summarize")
			dirPath, _ := cmd.Flags().GetString("input")
			json, _ := cmd.Flags().GetString("json")

			scanner := exifsort.NewScanner()
			err := scanner.ScanDir(dirPath, ioWriter(quiet))
			if err != nil {
				fmt.Printf("Scan error %s\n", err.Error())
				return
			}

			if summarize {
				scanSummary(&scanner)
			}
			scanSave(&scanner, json)
		},
	}

	var scanFlags = []cmdStringFlag{
		{"i", "input", true, "Input Directory to scan media."},
		{"j", "json", false, "json file to save output to."},
	}

	scanCmd.Flags().BoolP("quiet", "q", false,
		"Suppress line by line printing.")
	scanCmd.Flags().BoolP("summarize", "s", false,
		"Print a summary of stats when done.")

	setStringFlags(scanCmd, scanFlags)

	return scanCmd
}
