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

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newRootCmd() *cobra.Command {
	// rootCmd represents the base command when called without any subcommands.
	return &cobra.Command{
		Use:   "exifsort",
		Short: "Sorting media by date using the exif information",
		Long: `exifsort sorts media in nested directories 
by primarily by accessing the exif information. 

The pipeline of commands are:
exifsort scan <src>
exifsort sort <action> <method> <src> <dst>
exifsort merge <src> dst>

Check out github.com/matchstick/exifsort for more details.
. `,
	}
}

const exitErr = 1

func Execute() {
	cobra.OnInitialize(initConfig)

	rootCmd := newRootCmd()

	rootCmd.AddCommand(newEvalCmd())
	rootCmd.AddCommand(newFilterCmd())
	rootCmd.AddCommand(newMergeCmd())
	rootCmd.AddCommand(newScanCmd())
	rootCmd.AddCommand(newSortCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(exitErr)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var cfgFile string
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(exitErr)
		}

		// Search config in home directory with name ".exifsort"
		// (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".exifsort")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
