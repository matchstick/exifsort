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

type sortCmd struct {
	src       string
	dst       string
	method    int
	action    int
	quiet     bool
	summarize bool
	cobraCmd  *cobra.Command
}

func (s *sortCmd) sortSummary(w *exifsort.WalkState, sorter *exifsort.Sorter) {
	fmt.Printf("Sorted Valid: %d\n", w.Valid())
	fmt.Printf("Sorted Invalid: %d\n", w.Invalid())
	fmt.Printf("Sorted Skipped: %d\n", w.Skipped())
	fmt.Printf("Sorted Total: %d\n", w.Total())

	if w.Invalid() == 0 {
		fmt.Println("No Files caused Errors")
		return
	}

	fmt.Println("Walk Errors were:")

	for path, err := range w.Errors() {
		fmt.Printf("\t%s\n", w.ErrStr(path, err))
	}

	fmt.Println("Index Errors were:")

	for path, err := range sorter.IndexErrors() {
		fmt.Printf("\t%s\n", w.ErrStr(path, err))
	}

	fmt.Println("Transfer Errors were:")

	for path, err := range sorter.TransferErrors() {
		fmt.Printf("\t%s\n", w.ErrStr(path, err))
	}
}

func srcCheck(src string) bool {
	info, err := os.Stat(src)
	if err != nil || !info.IsDir() {
		fmt.Printf("Input Directory \"%s\" has error (%s)\n",
			src, err.Error())
		return false
	}

	return true
}

func dstCheck(dst string) bool {
	// dst must not be created yet
	_, err := os.Stat(dst)
	if err == nil || os.IsExist(err) {
		fmt.Printf("Output directory \"%s\" must not exist\n", dst)
		return false
	}

	return true
}

func (s *sortCmd) sortLongHelp() string {
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

// Here we finally do the work.
func (s *sortCmd) sortExecute() {
	writer := ioWriter(s.quiet)
	// Here we walk the directory and get stats
	w := exifsort.ScanDir(s.src, writer)

	// Now we take those stats and Sort them.
	sorter, err := exifsort.NewSorter(w, s.method)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	// Transfer the files to the dst
	err = sorter.Transfer(s.dst, s.action, writer)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	if s.summarize {
		s.sortSummary(&w, sorter)
	}
}

func newCobraCmd(s *sortCmd) *cobra.Command {
	return &cobra.Command{
		Use:   "sort",
		Short: "Accepts an input directory and will sort media by time created",
		// Very long help message so we moved it to a func.
		Long: s.sortLongHelp(),
		Args: cobra.MinimumNArgs(numSortArgs),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			s.quiet, _ = cmd.Flags().GetBool("quiet")
			s.summarize, _ = cmd.Flags().GetBool("summarize")

			s.src = args[0]
			s.dst = args[1]
			methodArg := args[2]
			actionArg := args[3]

			if !srcCheck(s.src) {
				return
			}

			if !dstCheck(s.dst) {
				return
			}

			s.method, err = exifsort.ParseMethod(methodArg)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}

			s.action, err = exifsort.ParseAction(actionArg)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}

			s.sortExecute()
		},
	}
}

func newSortCmd() *cobra.Command {
	// sortCmd represents the sort command.
	var cmd sortCmd
	cmd.cobraCmd = newCobraCmd(&cmd)

	cmd.cobraCmd.Flags().BoolP("quiet", "q", false,
		"Suppress line by line printing.")
	cmd.cobraCmd.Flags().BoolP("summarize", "s", false,
		"Print a summary of stats when done.")

	return cmd.cobraCmd
}
