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
	Short: "Scan Directory for Exif Dates",
	Long: `Scan a directory for Exif Date Info. Has two modes: 

	'line'    - a line for every file found and scanned
	'summary' - a compact summary of what was found 

Usage: exifSort scan <dir> -mode=[line|summary]

`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		dirPath := args[0]
		info, err := os.Stat(dirPath)
		if err != nil {
			fmt.Printf("Error with directory: %s\n", err.Error())
			return
		}
		if info.IsDir() == false {
			fmt.Print("Scan requires a directory as an argument\n")
			return
		}
		exifSort.ScanDir(dirPath)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
