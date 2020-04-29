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

func fileReadable(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	defer file.Close()
	if err != nil {
		return err
	}
	return nil
}

// evalCmd represents the eval command
var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Evals exif date data for one file only",
	Long: `Usage: exifSort eval <filename>
retreives the date data for one file from it's exif data. `,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		err := fileReadable(filePath)
		if err != nil {
			fmt.Printf("%q\n", err)
			return
		}
		entry, err := exifSort.ExtractExifDate(filePath)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		if entry.Valid == false {
			fmt.Printf("None\n")
			return
		}
		t := entry.Time
		dateStr := fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d",
					t.Year(), t.Month(), t.Day(),
					t.Hour(), t.Minute(), t.Second())
		fmt.Printf("%s\n", dateStr)
	},
}

func init() {
	rootCmd.AddCommand(evalCmd)
}
