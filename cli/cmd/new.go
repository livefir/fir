/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/livefir/fir/cli/entgo"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new fir project",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		projectName := "quickstart"
		if len(args) > 0 {
			projectName = args[0]
		}
		entgo.New(projectName)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
