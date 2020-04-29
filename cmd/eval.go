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
	"strings"
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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Eval: " + strings.Join(args, " "))
		filePath := args[0]
		err := fileReadable(filePath)
		if err != nil {
			fmt.Printf("%q\n", err)
			return
		}
		entry, err := exifSort.ExtractExifDate(filePath)
		if err != nil {
			fmt.Printf("%q\n", err)
			return
		}
		if entry.Valid == false {
			fmt.Printf("No Exif Data\n")
			return
		}
		fmt.Printf("Retrieved %+v\n", entry)
	},
}

func init() {
	rootCmd.AddCommand(evalCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// evalCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// evalCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
