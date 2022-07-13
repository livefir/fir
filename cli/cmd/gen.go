/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/schema/field"
	"github.com/adnaan/fir/cli/entgo"
)

var projectPath string
var pkg string

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate fir views and ent models from entgo.io schema",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		schemaPath, err := filepath.Abs(filepath.Join(projectPath, "schema"))
		if err != nil {
			log.Printf("schema path error: %v", err)
			cmd.Help()
			return
		}
		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			log.Printf("no schema directory found at path: %v\n", schemaPath)
			cmd.Help()
			return
		}
		modelsPkg := filepath.Join(pkg, "models")
		entgo.Generate(projectPath, modelsPkg)
		err = entc.Generate(schemaPath, &gen.Config{
			Header: `
			// Code generated (@generated) by entc, DO NOT EDIT.
		`,
			IDType:  &field.TypeInfo{Type: field.TypeInt},
			Target:  filepath.Join(projectPath, "models"),
			Package: modelsPkg,
		})
		if err != nil {
			log.Fatal("running ent codegen:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(genCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// genCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// genCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	wd, _ := os.Getwd()

	genCmd.Flags().StringVarP(&projectPath, "project", "p", wd, "path to project")
	genCmd.Flags().StringVarP(&pkg, "package", "P", "github.com/adnaan/fir/cli/testdata/todos", "project package path")
	genCmd.MarkFlagRequired("package")
}
