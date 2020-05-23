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

func fileReadable(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Evals exif date data for files",
	Long: `Report time for files not directories

	exifsort eval <files>...

	ARGUMENTS

	files
	file list (expanded by shell) that will have their exifDate reported`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, path := range args {
			err := fileReadable(path)
			if err != nil {
				fmt.Printf("%s, %q\n", path, err)
				continue
			}
			timeStr, err := exifsort.ExtractTimeStr(path)
			if err != nil {
				fmt.Printf("%s, %s\n", path, err)
				continue
			}
			fmt.Printf("%s, %s\n", path, timeStr)
		}
	},
}

func init() {
	rootCmd.AddCommand(evalCmd)
}
