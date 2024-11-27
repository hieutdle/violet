/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/hieutdle/violet/pkg/open"
	"github.com/spf13/cobra"
)

// openCmd represents the open command
var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open a website from the command line",
	RunE: func(cmd *cobra.Command, args []string) error {
		var id string
		if len(args) > 0 {
			id = args[0]
		}

		if id == "" {
			return fmt.Errorf("please provide a website ID")
		}

		// Open the website with the given ID
		if err := open.OpenWebsite(id); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(openCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// openCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// openCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
