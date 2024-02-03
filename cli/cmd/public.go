package cmd

import (
	"fmt"
	"os"

	"github.com/livefir/fir/gen"
	"github.com/spf13/cobra"
)

var (
	inDir      string
	outDir     string
	extensions []string
)

// publicCmd represents the public command
var publicCmd = &cobra.Command{
	Use:   "public",
	Short: "Generates the public folder containing the html files",
	Long: `The public command generates a public folder containing the html files in the project.
	It preserves the paths of the html files enabling a flexible project structure. The generated public directory
	can be embedded in the binary as is.`,
	Run: func(cmd *cobra.Command, args []string) {
		var opts []gen.PublicDirOption
		if inDir != "" {
			opts = append(opts, gen.InDir(inDir))
		}

		if outDir != "" {
			opts = append(opts, gen.OutDir(outDir))
		}

		if len(extensions) != 0 {
			opts = append(opts, gen.PublicFileExtensions(extensions))
		}

		if err := gen.GeneratePublicDir(opts...); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(publicCmd)
	publicCmd.Flags().StringVarP(&inDir, "in", "i", "", "path to input directory which contains the html template files")
	publicCmd.Flags().StringVarP(&outDir, "out", "o", "", "path to output directory")
	publicCmd.Flags().StringSliceVarP(&extensions, "extensions", "x", nil, "comma separated list of template exatensions e.g. .html,.tmpl")
}
