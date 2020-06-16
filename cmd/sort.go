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
	src       string
	dst       string
	json      string
	method    int
	action    int
	quiet     bool
	summarize bool
	cobraCmd  *cobra.Command
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
	writer := ioWriter(s.quiet)

	scanner := exifsort.NewScanner()

	switch {
	case s.src != "":
		// Here we walk the directory and get stats
		err := scanner.ScanDir(s.src, writer)
		if err != nil {
			fmt.Printf("Input Directory \"%s\" has error (%s)\n",
				s.src, err.Error())
			return
		}
	case s.json != "":
		// Or we get stats from a json file
		err := scanner.Load(s.json)
		if err != nil {
			fmt.Printf("Load of \"%s\" has error (%s)\n",
				s.json, err.Error())
			return
		}
	default:
		fmt.Printf("Inputs were not chosen\n")
		return
	}

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

func (s *sortCmd) newMethodCmd(method string, action string) *cobra.Command {
	const numMethodCmdArgs = 2

	return &cobra.Command{
		Use:   method,
		Short: fmt.Sprintf("Need to add subcommand for %s %s", action, method),
		// Very long help message so we moved it to a func.
		Long: s.sortLongHelp(),
		Args: cobra.MinimumNArgs(numMethodCmdArgs),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			s.src = args[0]
			s.dst = args[1]
			s.json, _ = cmd.Flags().GetString("json")

			// User needs to set either json or src, but not both.

			// Has the user not set any inputs?
			if s.json == "" && s.src == "" {
				fmt.Printf("Must set input with either -j or -i.\n")
				return
			}

			// Has the user set both?
			if s.json != "" && s.src != "" {
				fmt.Printf("Cannot use both -j and -i for input.\n")
				return
			}

			// We create directory before executing.
			// It would not be cool to spend a lot of time
			// then fail due to perms or previous output
			// directory.
			err = outputCreate(s.dst)
			if err != nil {
				return
			}

			s.method, err = exifsort.ParseMethod(method)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}

			s.action, err = exifsort.ParseAction(action)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}

			s.sortExecute()
		},
	}
}

func (s *sortCmd) newOpCmd(action string) *cobra.Command {
	return &cobra.Command{
		Use:   action,
		Short: "Need to add usage for " + action + " subcommand.",
		// Very long help message so we moved it to a func.
		Long: s.sortLongHelp(),
	}
}

func newSortRootCmd(s *sortCmd) *cobra.Command {
	return &cobra.Command{
		Use:   "sort",
		Short: "Accepts an input directory and will sort media by time created",
		// Very long help message so we moved it to a func.
		Long: s.sortLongHelp(),
	}
}

func newSortCmd() *cobra.Command {
	var sortFlags = []cmdStringFlag{
		{"j", "json", false, "Json File input to load media."},
	}

	// sortCmd represents the sort command.
	var s sortCmd
	s.cobraCmd = newSortRootCmd(&s)

	setStringFlags(s.cobraCmd, sortFlags)
	copyCmd := s.newOpCmd("copy")
	moveCmd := s.newOpCmd("move")

	copyCmd.AddCommand(s.newMethodCmd("year", "copy"), s.newMethodCmd("month", "copy"),
		s.newMethodCmd("day", "copy"))

	moveCmd.AddCommand(s.newMethodCmd("year", "move"),
		s.newMethodCmd("month", "move"),
		s.newMethodCmd("day", "move"))

	s.cobraCmd.AddCommand(moveCmd, copyCmd)

	return s.cobraCmd
}
