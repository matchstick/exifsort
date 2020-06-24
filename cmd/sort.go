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

type sortCmd struct {
	src      string
	dst      string
	method   exifsort.Method
	action   exifsort.Action
	cobraCmd *cobra.Command
}

func (s *sortCmd) sortSummary(scanner *exifsort.Scanner,
	sorter *exifsort.Sorter) {
	scanSummary(scanner)

	if len(sorter.IndexErrors) != 0 {
		fmt.Println("## Index Errors were:")

		for path, err := range sorter.IndexErrors {
			fmt.Printf("## \t%s: (%s)\n", path, err)
		}
	}

	if len(sorter.TransferErrors) != 0 {
		fmt.Println("## Transfer Errors were:")

		for path, err := range sorter.TransferErrors {
			fmt.Printf("##\t%s: (%s)\n", path, err)
		}
	}
}

func outputCreate(dst string) error {
	err := os.Mkdir(dst, 0755)
	if err != nil {
		return err
	}

	return nil
}

func (s *sortCmd) isSrcDir() bool {
	info, err := os.Stat(s.src)

	// Cannot even stat it
	if err != nil {
		return false
	}

	return info.IsDir()
}

func (s *sortCmd) sortLongHelp() string {
	return `Sort directory by Exif Date Info. 

	exifsort sort <action> <method> <src> <dst>

	sort command performs a number of steps:

	1. Collect media information via scanning a directory or reading a json file from scan
	2. Indexing the media by method
	3. Create a directory for output
	4. Transfer media to the output structed and sorted.

	ARGUMENTS

	action
	Choice of how to move files from src to dst.
	Valid values are 'copy' or 'move'

	method
	Choice of how to index the media in the new directory.
	Valid values are 'year', 'month' or 'day'.

	src
	directory or json file to receive media to sort

	dst
	directory to create to transfer media
	`
}

// Here we finally do the work.
func (s *sortCmd) sortExecute() {
	scanner := exifsort.NewScanner()

	var err error
	if s.isSrcDir() {
		// Here we walk the directory and get stats
		err = scanner.ScanDir(s.src, os.Stdout)
	} else {
		// Or we get stats from a json file
		err = scanner.Load(s.src)
	}

	if err != nil {
		fmt.Printf("\"%s\" error (%s)\n", s.src, err.Error())
		return
	}

	// Now we ke those stats and Sort them.
	sorter, err := exifsort.NewSorter(scanner, s.method)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	// Transfer the files to the dst
	err = sorter.Transfer(s.dst, s.action, os.Stdout)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	s.sortSummary(&scanner, sorter)
}

func (s *sortCmd) newSortMethodCmd(action exifsort.Action,
								   method exifsort.Method) *cobra.Command {
	const numMethodCmdArgs = 2

	methodStr := method.String()
	actionStr := action.String()

	return &cobra.Command{
		Use:   methodStr,
		Short: fmt.Sprintf("Transfer by %s then sort by %s", actionStr, methodStr),
		// Very long help message so we moved it to a func.
		Long: s.sortLongHelp(),
		Args: cobra.MinimumNArgs(numMethodCmdArgs),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			s.src = args[0]
			s.dst = args[1]
			s.method = method
			s.action = action

			// We create directory before executing.
			// It would not be cool to spend a lot of time
			// then fail due to perms or previous output
			// directory.
			err = outputCreate(s.dst)
			if err != nil {
				return
			}

			s.sortExecute()
		},
	}
}

func (s *sortCmd) newSortActionCmd(action exifsort.Action) *cobra.Command {
	actionStr := action.String()

	actionCmd := &cobra.Command{
		Use:   actionStr,
		Short: "Transfer by " + actionStr,
		// Very long help message so we moved it to a func.
		Long: s.sortLongHelp(),
	}

	for _, method := range exifsort.Methods() {
		methodCmd := s.newSortMethodCmd(action, method)
		actionCmd.AddCommand(methodCmd)
	}

	return actionCmd
}

func newSortRootCmd(s *sortCmd) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "sort",
		Short: "Accepts an input directory and will sort media by time created",
		// Very long help message so we moved it to a func.
		Long: s.sortLongHelp(),
	}

	for _, action := range exifsort.Actions() {
		actionCmd := s.newSortActionCmd(action)
		rootCmd.AddCommand(actionCmd)
	}

	return rootCmd
}

func newSortCmd() *cobra.Command {
	// sortCmd represents the sort command.
	var s sortCmd
	s.cobraCmd = newSortRootCmd(&s)

	return s.cobraCmd
}
