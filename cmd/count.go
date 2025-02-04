package cmd

import (
	"fmt"

	"github.com/hieutdle/violet/pkg/count"
	"github.com/spf13/cobra"
)

// countCmd represents the open command
var countCmd = &cobra.Command{
	Use:   "count",
	Short: "Count number of lines in a git repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(count.CountGitLines())
		return nil
	},
}

func init() {
	RootCmd.AddCommand(countCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// openCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// openCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
