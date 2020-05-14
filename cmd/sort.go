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

// sortCmd represents the sort command
var sortCmd = &cobra.Command{
	Use:   "sort",
	Short: "Accepts an input directory and wil move or copy all media files into an oputput directory sorted by time taken",
	Long: `sort takes in four arguments:
		srcDir: input directory of media files
		dstDir: directory it will create to output files in time sorted format
		method: "Year", "Month", "Day" - How to sort them by year, month or day
		action: "Copy", "Move" - Whether to copy or move files`,
	Args: cobra.MinimumNArgs(4),
	Run: func(cmd *cobra.Command, args []string) {

		quiet, _ := cmd.Flags().GetBool("quiet")
		summarize, _ := cmd.Flags().GetBool("summarize")

		srcDir := args[0]
		dstDir := args[1]
		methodArg := args[2]
		actionArg := args[3]

		info, err := os.Stat(srcDir)
		if err != nil || info.IsDir() == false {
			fmt.Printf("Input Directory \"%s\" has error (%s)\n", srcDir, err.Error())
			return
		}
		// dstDir must not be created yet
		_, err = os.Stat(dstDir)
		if err == nil || os.IsExist(err) {
			fmt.Printf("Output directory \"%s\" must not exist\n", dstDir)
			return
		}
		method, err := exifSort.MethodParse(methodArg)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			return
		}
		action, err := exifSort.ActionParse(actionArg)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			return
		}
		exifSort.SortDir(srcDir, dstDir, method, action, summarize, !quiet)
	},
}

func init() {
	rootCmd.AddCommand(sortCmd)
	sortCmd.Flags().BoolP("quiet", "q", false,
		"Don't print output while scanning")
	sortCmd.Flags().BoolP("summarize", "s", false,
		"Print a summary when done scanning")
}
