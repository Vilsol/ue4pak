package cmd

import (
	"fmt"
	"github.com/Vilsol/ue4pak/parser"
	"github.com/fatih/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test parse the provided paks",
	Run: func(cmd *cobra.Command, args []string) {
		color.NoColor = false

		paks, err := filepath.Glob(cmd.Flag("pak").Value.String())

		if err != nil {
			panic(err)
		}

		for _, f := range paks {
			fmt.Println("Parsing file:", f)

			file, err := os.OpenFile(f, os.O_RDONLY, 0644)

			if err != nil {
				panic(err)
			}

			pak := parser.Parse(file)

			summaries := make(map[string]*parser.FPackageFileSummary, 0)

			// First pass, parse summaries
			for j, record := range pak.Index.Records {
				trimmed := trim(record.FileName)
				if strings.HasSuffix(trimmed, "uasset") {
					fmt.Printf("Reading Record: %d: %#v\n", j, record)
					summaries[trimmed[0:strings.Index(trimmed, ".uasset")]] = record.ReadUAsset(file)

					/*
						fmt.Println("\nNames:")
						for i, name := range lastUasset.Names {
							fmt.Printf("%#x: %s\n", i, name.Name)
						}

						fmt.Println("\nImports:")
						for i, name := range lastUasset.Imports {
							fmt.Printf("%#x: %#v\n", i, name)
						}

						fmt.Println("\nExports:")
						for i, name := range lastUasset.Exports {
							fmt.Printf("%#x: %#v\n", i, name)
						}
					*/
				}
			}

			// Second pass, parse exports
			for j, record := range pak.Index.Records {
				trimmed := trim(record.FileName)
				if strings.HasSuffix(trimmed, "uexp") {
					summary, ok := summaries[trimmed[0:strings.Index(trimmed, ".uexp")]]

					if !ok {
						fmt.Printf("Unable to read record. Missing uasset: %d: %#v\n", j, record)
						continue
					}

					fmt.Printf("Reading Record: %d: %#v\n", j, record)

					record.ReadUExp(file, summary)
					/*
						uexp := record.ReadUExp(file, summary)

						for export, properties := range uexp {
							if export.TemplateIndex.Reference != nil {
								if imp, ok := export.TemplateIndex.Reference.(*parser.FObjectImport); ok {
									switch trim(imp.ClassName) {
									case "FGFactoryConnectionComponent":
										fallthrough
									case "FGBuildSubCategory":
										fallthrough
									case "FGBuildingDescriptor":
										fallthrough
									case "FGBuildableStorage":
										fallthrough
									case "FGRecipe":
										fallthrough
									case "FGSchematic":
										fmt.Println()
										fmt.Printf("%#v\n", export.TemplateIndex.Reference)

										fmt.Println()

										for _, property := range properties {
											fmt.Printf("%s [%v]: %#v\n", trim(property.Name), property.TagData, property.Tag)
										}

										fmt.Println()

										break
									}
								}
							}
						}
					*/
				}
			}
		}
	},
}

func trim(s string) string {
	return strings.Trim(s, "\x00")
}
