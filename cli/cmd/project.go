package cmd

import (
	"github.com/livefir/fir/cli/gen"
	"github.com/spf13/cobra"
)

var (
	projectName string
)

// projectCmd represents the new command
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Creates a new fir project",
	Long:  `Creates a new fir project`,
	Run: func(cmd *cobra.Command, args []string) {
		gen.NewProject(projectName)
	},
}

func init() {
	projectCmd.Flags().StringVarP(&projectName, "name", "n", "quickstart", "name of the project")
}
