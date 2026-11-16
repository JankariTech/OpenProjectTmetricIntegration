/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "validate the data",
	Long:  `call with 'tmetric' or 'openproject'`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("please supply a subcommand (tmetric or openproject)")
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
