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

func (s *sortCmd) sortSummary(scanner *exifsort.Scanner,
	sorter *exifsort.Sorter) {
	fmt.Printf("Sorted Valid: %d\n", scanner.NumValid())
	fmt.Printf("Sorted Invalid: %d\n", scanner.NumInvalid())
	fmt.Printf("Sorted Skipped: %d\n", scanner.NumSkipped())
	fmt.Printf("Sorted Total: %d\n", scanner.NumTotal())

	if scanner.NumInvalid() == 0 {
		fmt.Println("No Files caused Errors")
		return
	}

	fmt.Println("Walk Errors were:")

	for path, err := range scanner.Errors {
		fmt.Printf("\t%s\n", scanner.ErrStr(path, err))
	}

	fmt.Println("Index Errors were:")

	for path, err := range sorter.IndexErrors {
		fmt.Printf("\t%s\n", scanner.ErrStr(path, err))
	}

	fmt.Println("Transfer Errors were:")

	for path, err := range sorter.TransferErrors {
		fmt.Printf("\t%s\n", scanner.ErrStr(path, err))
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

	exifsort sort <options>

	exifsort will recursively check every file in an input directory and
	then create antoher directory structure organized by time to either
	move or copy the files into
	`
}

// Here we finally do the work.
func (s *sortCmd) sortExecute() {
	writer := ioWriter(s.quiet)
	// Here we walk the directory and get stats
	scanner := exifsort.NewScanner()
	scanner.ScanDir(s.src, writer)

	// Now we take those stats and Sort them.
	sorter, err := exifsort.NewSorter(scanner, s.method)
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
		s.sortSummary(&scanner, sorter)
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

			s.src, _ = cmd.Flags().GetString("input")
			s.dst, _ = cmd.Flags().GetString("output")
			methodArg, _ := cmd.Flags().GetString("method")
			actionArg, _ := cmd.Flags().GetString("action")

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
	var sortFlags = []cmdStringFlag{
		{"i", "input", "Input Directory to scan media."},
		{"o", "output", "Output Directory to transfer media. (Must not exist.)"},
		{"m", "method", "Method to index media in output directory. <year|month|day>"},
		{"a", "action", "Transfer Action: <copy|move>"},
	}

	// sortCmd represents the sort command.
	var cmd sortCmd
	cmd.cobraCmd = newCobraCmd(&cmd)

	cmd.cobraCmd.Flags().BoolP("quiet", "q", false,
		"Suppress line by line printing.")
	cmd.cobraCmd.Flags().BoolP("summarize", "s", false,
		"Print a summary of stats when done.")

	setRequiredFlags(cmd.cobraCmd, sortFlags)

	return cmd.cobraCmd
}
