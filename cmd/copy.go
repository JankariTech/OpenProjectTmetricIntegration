/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy time entries from Tmetric to OpenProject",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("copy called")
		// get all time entries from tmetric
		// filter out those that have the tag 'transferred-to-openproject' set
		// check if any entry has no work-type assigned or is not linked to an openproject WP, refuse to copy in that case
		// find all work-types and check if there are matching values in openproject, refuse to copy if invalid work-types are found
		// if all entries are valid:
		// for each time entry
		// copy the time entry to openproject
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// copyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// copyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
