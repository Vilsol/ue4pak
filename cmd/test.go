package cmd

import (
	"fmt"
	"github.com/Vilsol/ue4pak/parser"
	"github.com/fatih/color"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "test",
	Short: "Test parse the provided pak file",
	Run: func(cmd *cobra.Command, args []string) {
		color.NoColor = false

		file, err := os.OpenFile(cmd.Flag("pak").Value.String(), os.O_RDONLY, 0644)

		if err != nil {
			panic(err)
		}

		pak := parser.Parse(file)

		var lastUasset *parser.FPackageFileSummary

		for _, record := range pak.Index.Records {
			// fmt.Printf("Reading Record: %d: %#v\n", j, record)

			trimmed := strings.Trim(record.FileName, "\x00")
			if strings.HasSuffix(trimmed, "uasset") {
				lastUasset = record.ReadUAsset(file)

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
			} else if strings.HasSuffix(trimmed, "uexp") {
				uexp := record.ReadUExp(file, lastUasset)

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
			}
		}
	},
}

func trim(s string) string {
	return strings.Trim(s, "\x00")
}
