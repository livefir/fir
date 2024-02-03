package cmd

import (
	"github.com/livefir/fir/cli/gen"
	"github.com/spf13/cobra"
)

var (
	routeName string
)

// routeCmd represents the new command
var routeCmd = &cobra.Command{
	Use:   "route",
	Short: "Creates a new fir route",
	Long:  `Creates a new fir route`,
	Run: func(cmd *cobra.Command, args []string) {
		gen.NewRoute(routeName)
	},
}

func init() {
	routeCmd.Flags().StringVarP(&routeName, "name", "n", "index", "route name")
}
