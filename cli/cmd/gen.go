package cmd

import "github.com/spf13/cobra"

var gendCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generates fir bolilerplate",
	Long:  `The gen command generates files and directories for a fir project. It has subcommands.`,
}

func init() {
	rootCmd.AddCommand(gendCmd)
	gendCmd.AddCommand(projectCmd)
	gendCmd.AddCommand(routeCmd)
}
