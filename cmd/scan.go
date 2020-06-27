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

func scanSummary(s *exifsort.Scanner) {
	// If we are not walking a directory there is no summary
	if s.Input != exifsort.ScannerInputDir {
		return
	}

	fmt.Printf("## Scanned Total: %d\n", s.NumTotal())
	fmt.Printf("## Scanned Skipped: %d\n", s.SkippedCount)
	fmt.Printf("## Scanned Data: %d\n", len(s.Data))

	for extension, num := range s.NumDataTypes {
		fmt.Printf("##\t [%s]: %d\n", extension, num)
	}

	if len(s.ScanErrors) != 0 {
		fmt.Println("## Scanned Errors were:")

		for path, err := range s.ScanErrors {
			fmt.Printf("##\t%s: (%s)\n", path, err)
		}
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

	exifsort scan <src> [--json <file>] 

	ARGUMENTS

	src 
	directory to scan for media date informaiton.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dirPath := args[0]
			json, _ := cmd.Flags().GetString("json")

			scanner := exifsort.NewScanner()
			err := scanner.ScanDir(dirPath, os.Stdout)
			if err != nil {
				fmt.Printf("Scan error %s\n", err.Error())
				return
			}

			scanSummary(&scanner)
			scanSave(&scanner, json)
		},
	}

	var scanFlags = []cmdStringFlag{
		{"j", "json", false, "json file to save output to."},
	}

	setStringFlags(scanCmd, scanFlags)

	return scanCmd
}
