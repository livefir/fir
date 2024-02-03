/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/livefir/fir/cli/gen"
	"github.com/spf13/cobra"
)

var (
	projectName string
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new [projectName]",
	Short: "Create a new fir project",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		projectName := "quickstart"
		if len(args) > 0 {
			projectName = args[0]
		}
		gen.NewQuickstart(projectName)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	rootCmd.Flags().StringVarP(&projectName, "project", "p", "quickstart", "name of the project")
}
