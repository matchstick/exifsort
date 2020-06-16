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
	method   int
	action   int
	cobraCmd *cobra.Command
}

func (s *sortCmd) sortSummary(scanner *exifsort.Scanner,
	sorter *exifsort.Sorter) {
	if s.src != "" {
		scanSummary(scanner)
	}

	if len(sorter.IndexErrors) != 0 {
		fmt.Println("Index Errors were:")

		for path, err := range sorter.IndexErrors {
			fmt.Printf("\t%s: (%s)\n", path, err)
		}
	}

	if len(sorter.TransferErrors) != 0 {
		fmt.Println("Transfer Errors were:")

		for path, err := range sorter.TransferErrors {
			fmt.Printf("\t%s: (%s)\n", path, err)
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

	exifsort sort <copy | move> <year | month | day> <src> <dst> [--json <json>]

	exifsort will recursively check every file in an input directory and
	then create antoher directory structure organized by time to either
	move or copy the files into
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

func (s *sortCmd) newMethodCmd(action int, method int) *cobra.Command {
	const numMethodCmdArgs = 2

	methodStr := exifsort.MethodMap()[method]
	actionStr := exifsort.ActionMap()[action]

	return &cobra.Command{
		Use:   methodStr,
		Short: fmt.Sprintf("Need to add subcommand for %s %s", actionStr, methodStr),
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

func (s *sortCmd) newActionCmd(action int) *cobra.Command {
	actionStr := exifsort.ActionMap()[action]

	actionCmd := &cobra.Command{
		Use:   actionStr,
		Short: "Need to add usage for " + actionStr + " subcommand.",
		// Very long help message so we moved it to a func.
		Long: s.sortLongHelp(),
	}

	for method := range exifsort.MethodMap() {
		methodCmd := s.newMethodCmd(action, method)
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

	for action := range exifsort.ActionMap() {
		actionCmd := s.newActionCmd(action)
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
