/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/schema/field"
	"github.com/adnaan/fir/cli/entgo"
	"golang.org/x/mod/modfile"
)

var projectPath string
var module string

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

		goModPath := filepath.Join(projectPath, "go.mod")
		if _, err := os.Stat(goModPath); !os.IsNotExist(err) {
			goModBytes, err := ioutil.ReadFile(goModPath)
			if err != nil {
				log.Printf("error reading %s: %v\n.", goModPath, err)
				return
			}
			module = modfile.ModulePath(goModBytes)
		}

		if module == "" {
			log.Printf("module not set and go.mod doesn't exist %s\n", goModPath)
			cmd.Help()
			return
		}

		fmt.Println("using module:", module)

		modelsPkg := filepath.Join(module, "models")
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

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	genCmd.Flags().StringVarP(&projectPath, "project", "p", wd, "path to project")
	genCmd.Flags().StringVarP(&module, "module", "m", "", "module name(go.mod) or package(github.com/x/y/z) of the current project")
}
