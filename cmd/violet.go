/*
Copyright © 2024 Hieu Le <hieu.tdle@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = rootCmd()

func init() {
	versionTemplate := `{{printf "%s: %s - version %s\n" .Name .Short .Version}}`
	RootCmd.SetVersionTemplate(versionTemplate)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.violet.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func rootCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "violet",
		Short:         "My personal CLI",
		SilenceErrors: false,
		SilenceUsage:  true,
		Version:       "0.3.0",
	}
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
