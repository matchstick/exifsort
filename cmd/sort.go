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
			fmt.Printf("\t%s\n", exifsort.ErrStr(path, err))
		}
	}

	if len(sorter.TransferErrors) != 0 {
		fmt.Println("Transfer Errors were:")

		for path, err := range sorter.TransferErrors {
			fmt.Printf("\t%s\n", exifsort.ErrStr(path, err))
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

	exifsort sort <options>

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

func newCobraCmd(s *sortCmd) *cobra.Command {
	return &cobra.Command{
		Use:   "sort",
		Short: "Accepts an input directory and will sort media by time created",
		// Very long help message so we moved it to a func.
		Long: s.sortLongHelp(),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			s.quiet, _ = cmd.Flags().GetBool("quiet")
			s.summarize, _ = cmd.Flags().GetBool("summarize")

			s.src, _ = cmd.Flags().GetString("input")
			s.dst, _ = cmd.Flags().GetString("output")
			s.json, _ = cmd.Flags().GetString("json")
			methodArg, _ := cmd.Flags().GetString("method")
			actionArg, _ := cmd.Flags().GetString("action")

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
		{"a", "action", true,
			"Transfer Action: <copy|move>"},
		{"i", "input", false,
			"Input Directory to scan media."},
		{"j", "json", false,
			"Json File input to load media."},
		{"m", "method", true,
			"Method to index media in output directory. <year|month|day>"},
		{"o", "output", true,
			"Output Directory to transfer media. (Must not exist.)"},
	}

	// sortCmd represents the sort command.
	var cmd sortCmd
	cmd.cobraCmd = newCobraCmd(&cmd)

	setStringFlags(cmd.cobraCmd, sortFlags)
	addCommonFlags(cmd.cobraCmd)

	return cmd.cobraCmd
}
