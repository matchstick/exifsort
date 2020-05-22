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
	"os"

	exifsort "github.com/matchstick/exifsort/lib"
	"github.com/spf13/cobra"
)

// sortCmd represents the sort command
var sortCmd = &cobra.Command{
	Use:   "sort",
	Short: "Accepts an input directory and will sort media by time created",
	Long: `Sort directory by Exif Date Info. 

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
	How the media is transferred from src to dst`,
	Args: cobra.MinimumNArgs(4),
	Run: func(cmd *cobra.Command, args []string) {

		quiet, _ := cmd.Flags().GetBool("quiet")
		summarize, _ := cmd.Flags().GetBool("summarize")

		srcDir := args[0]
		dstDir := args[1]
		methodArg := args[2]
		actionArg := args[3]

		info, err := os.Stat(srcDir)
		if err != nil || !info.IsDir() {
			fmt.Printf("Input Directory \"%s\" has error (%s)\n", srcDir, err.Error())
			return
		}
		// dstDir must not be created yet
		_, err = os.Stat(dstDir)
		if err == nil || os.IsExist(err) {
			fmt.Printf("Output directory \"%s\" must not exist\n", dstDir)
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
		err = exifsort.SortDir(srcDir, dstDir, method, action, summarize, !quiet)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(sortCmd)
	sortCmd.Flags().BoolP("quiet", "q", false,
		"Suppress line by line printing.")
	sortCmd.Flags().BoolP("summarize", "s", false,
		"Print a summary of stats when done.")
}
