package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/Vilsol/ue4pak/parser"
	"github.com/fatih/color"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
)

func init() {
	format = classTreeCmd.Flags().StringP("format", "f", "json", "Output format type")
	output = classTreeCmd.Flags().StringP("output", "o", "extracted.json", "Output file")
	pretty = classTreeCmd.Flags().Bool("pretty", false, "Whether to output in a pretty format")

	classTreeCmd.MarkFlagRequired("assets")

	rootCmd.AddCommand(classTreeCmd)
}

var classTreeCmd = &cobra.Command{
	Use:   "class-tree",
	Short: "Read paks and output their class trees",
	Run: func(cmd *cobra.Command, args []string) {
		color.NoColor = false

		paks, err := filepath.Glob(cmd.Flag("pak").Value.String())

		if err != nil {
			panic(err)
		}

		patterns := make([]glob.Glob, len(*assets))
		for i, asset := range *assets {
			patterns[i] = glob.MustCompile(asset)
		}

		results := make(map[string]map[string]string, 0)

		open, err := os.OpenFile("much-data.txt", os.O_WRONLY|os.O_CREATE, 0644)

		fmt.Println(err)

		for _, f := range paks {
			fmt.Println("Parsing file:", f)

			file, err := os.OpenFile(f, os.O_RDONLY, 0644)

			if err != nil {
				panic(err)
			}

			p := parser.NewParser(file)
			things, _ := p.ProcessPak(nil)

			for _, thing := range things {
				for _, export := range thing.Exports {
					open.WriteString(fmt.Sprintf("Class: %s%s\n", trim(export.Export.ObjectName), BuildClassTree(export.Export.ClassIndex)))
					open.WriteString(fmt.Sprintf("Super: %s%s\n", trim(export.Export.ObjectName), BuildSuperTree(export.Export.SuperIndex)))
					open.WriteString(fmt.Sprintf("Templ: %s%s\n", trim(export.Export.ObjectName), BuildTemplateTree(export.Export.TemplateIndex)))
					open.WriteString(fmt.Sprintf("Outer: %s%s\n", trim(export.Export.ObjectName), BuildOuterTree(export.Export.OuterIndex)))
				}
			}

			// indent, _ := json.MarshalIndent(concreteRecipe.Exports, "", " ")
			// fmt.Println(string(indent))
			// fmt.Printf("%#v\n", concreteRecipe.ExportRecord.FileName)
		}

		open.Close()

		var resultBytes []byte

		if *format == "json" {
			if *pretty {
				resultBytes, err = json.MarshalIndent(results, "", "  ")
			} else {
				resultBytes, err = json.Marshal(results)
			}

			if err != nil {
				panic(err)
			}
		} else {
			panic("Unknown output format: " + *format)
		}

		err = ioutil.WriteFile(*output, resultBytes, 0644)

		if err != nil {
			panic(err)
		}
	},
}

func BuildClassTree(index *parser.FPackageIndex) string {
	result := ""

	indexRef := index.Reference
	if indexRef == nil {
		result += " -> ROOT"
	} else if ref, ok := indexRef.(*parser.FObjectImport); ok {
		if ref != nil {
			result += fmt.Sprintf(" -> [I] %s", trim(ref.ObjectName))
		} else {
			result += " -> ROOT"
		}
	} else if ref, ok := indexRef.(*parser.FObjectExport); ok {
		if ref != nil {
			result += fmt.Sprintf(" -> [E] %s", trim(ref.ObjectName))
			result += BuildClassTree(ref.ClassIndex)
		} else {
			result += " -> ROOT"
		}
	} else {
		result += " -> UNKNOWN???"
	}

	return result
}

func BuildSuperTree(index *parser.FPackageIndex) string {
	result := ""

	indexRef := index.Reference
	if indexRef == nil {
		result += " -> ROOT"
	} else if ref, ok := indexRef.(*parser.FObjectImport); ok {
		if ref != nil {
			result += fmt.Sprintf(" -> [I] %s", trim(ref.ObjectName))
		} else {
			result += " -> ROOT"
		}
	} else if ref, ok := indexRef.(*parser.FObjectExport); ok {
		if ref != nil {
			result += fmt.Sprintf(" -> [E] %s", trim(ref.ObjectName))
			result += BuildSuperTree(ref.ClassIndex)
		} else {
			result += " -> ROOT"
		}
	} else {
		result += " -> UNKNOWN???"
	}

	return result
}

func BuildTemplateTree(index *parser.FPackageIndex) string {
	result := ""

	indexRef := index.Reference
	if indexRef == nil {
		result += " -> ROOT"
	} else if ref, ok := indexRef.(*parser.FObjectImport); ok {
		if ref != nil {
			result += fmt.Sprintf(" -> [I] %s", trim(ref.ObjectName))
		} else {
			result += " -> ROOT"
		}
	} else if ref, ok := indexRef.(*parser.FObjectExport); ok {
		if ref != nil {
			result += fmt.Sprintf(" -> [E] %s", trim(ref.ObjectName))
			result += BuildTemplateTree(ref.ClassIndex)
		} else {
			result += " -> ROOT"
		}
	} else {
		result += " -> UNKNOWN???"
	}

	return result
}

func BuildOuterTree(index *parser.FPackageIndex) string {
	result := ""

	indexRef := index.Reference
	if indexRef == nil {
		result += " -> ROOT"
	} else if ref, ok := indexRef.(*parser.FObjectImport); ok {
		if ref != nil {
			result += fmt.Sprintf(" -> [I] %s", trim(ref.ObjectName))
		} else {
			result += " -> ROOT"
		}
	} else if ref, ok := indexRef.(*parser.FObjectExport); ok {
		if ref != nil {
			result += fmt.Sprintf(" -> [E] %s", trim(ref.ObjectName))
			result += BuildOuterTree(ref.ClassIndex)
		} else {
			result += " -> ROOT"
		}
	} else {
		result += " -> UNKNOWN???"
	}

	return result
}
