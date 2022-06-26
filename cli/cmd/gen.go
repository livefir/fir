/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/schema/field"
	"github.com/adnaan/fir/cli/entgo"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate fir views and ent models from entgo.io schema",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		schema := "./testdata/todos/schema"
		viewsPath := "testdata/todos/views"
		modelsPath := "./testdata/todos/models"
		templateAssetsPath := "./template_assets"
		templatesPath := entgo.PrepareTemplates(schema, viewsPath, templateAssetsPath)
		err := entc.Generate(schema, &gen.Config{
			Header: `
			// Code generated (@generated) by entc, DO NOT EDIT.
		`,
			IDType:  &field.TypeInfo{Type: field.TypeInt},
			Target:  modelsPath,
			Package: "github.com/adnaan/fir/cli/testdata/todos/models",
		}, entc.Extensions(&entgo.ViewExtension{TemplatesPath: templatesPath}))
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
}
