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

const numSortArgs = 4

func sortSummary(w *exifsort.WalkState) {
	fmt.Printf("Sorted Valid: %d\n", w.Valid())
	fmt.Printf("Sorted Invalid: %d\n", w.Invalid())
	fmt.Printf("Sorted Skipped: %d\n", w.Skipped())
	fmt.Printf("Sorted Total: %d\n", w.Total())

	if w.Invalid() == 0 {
		fmt.Println("No Files caused Errors")
		return
	}

	fmt.Println("Walk Errors were:")

	for path, msg := range w.WalkErrs() {
		fmt.Printf("\t%s\n", w.ErrStr(path, msg))
	}

	fmt.Println("Transfer Errors were:")

	for path, msg := range w.TransferErrs() {
		fmt.Printf("\t%s\n", w.ErrStr(path, msg))
	}
}

func srcCheck(srcDir string) bool {
	info, err := os.Stat(srcDir)
	if err != nil || !info.IsDir() {
		fmt.Printf("Input Directory \"%s\" has error (%s)\n", srcDir, err.Error())
		return false
	}

	return true
}

func dstCheck(dstDir string) bool {
	// dstDir must not be created yet
	_, err := os.Stat(dstDir)
	if err == nil || os.IsExist(err) {
		fmt.Printf("Output directory \"%s\" must not exist\n", dstDir)
		return false
	}

	return true
}

func sortLongHelp() string {
	return `Sort directory by Exif Date Info. 

	exifsort sort [<options>...] <src> <dst> <method> <action>

	exifsort will recursively check every file in an input directory and
	then create antoher directory structure organized by time to either
	move or copy the files into

	ARGUMENTS

	src
	Input directory of media files

	dst
	Directory to create for output cannot exist

	method
	How to sort the media. It can be by "Year", "Month", or "Day"

		Year : dst -> year-> media
		Month: dst -> year-> month -> media
		Day  : dst -> year-> month -> day -> media

	action
	How the media is transferred from src to dst`
}

func newSortCmd() *cobra.Command {
	// sortCmd represents the sort command.
	var sortCmd = &cobra.Command{
		Use:   "sort",
		Short: "Accepts an input directory and will sort media by time created",
		// Very long help message so we moved it to a func.
		Long: sortLongHelp(),
		Args: cobra.MinimumNArgs(numSortArgs),
		Run: func(cmd *cobra.Command, args []string) {
			quiet, _ := cmd.Flags().GetBool("quiet")
			summarize, _ := cmd.Flags().GetBool("summarize")

			srcDir := args[0]
			dstDir := args[1]
			methodArg := args[2]
			actionArg := args[3]

			if !srcCheck(args[0]) {
				return
			}

			if !dstCheck(args[1]) {
				return
			}

			method, err := exifsort.ParseMethod(methodArg)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}
			action, err := exifsort.ParseAction(actionArg)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}
			w, err := exifsort.SortDir(srcDir, dstDir, method, action, ioWriter(quiet))
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}
			if summarize {
				sortSummary(&w)
			}
		},
	}

	sortCmd.Flags().BoolP("quiet", "q", false,
		"Suppress line by line printing.")
	sortCmd.Flags().BoolP("summarize", "s", false,
		"Print a summary of stats when done.")

	return sortCmd
}
