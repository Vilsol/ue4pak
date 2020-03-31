package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Vilsol/ue4pak/parser"
	"github.com/fatih/color"

	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

var assets *[]string

func init() {
	assets = extractCmd.Flags().StringSliceP("assets", "a", []string{}, "Comma-separated list of asset paths to extract. (supports glob) (required)")
	format = extractCmd.Flags().StringP("format", "f", "json", "Output format type")
	output = extractCmd.Flags().StringP("output", "o", "extracted.json", "Output file")
	pretty = extractCmd.Flags().Bool("pretty", false, "Whether to output in a pretty format")

	extractCmd.MarkFlagRequired("assets")

	rootCmd.AddCommand(extractCmd)
}

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract provided asset paths",
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

		results := make([]*parser.PakEntrySet, 0)

		for _, f := range paks {
			fmt.Println("Parsing file:", f)

			file, err := os.OpenFile(f, os.O_RDONLY, 0644)

			if err != nil {
				panic(err)
			}

			shouldProcess := func(name string) bool {
				for _, pattern := range patterns {
					if pattern.Match(name) {
						return true
					}
				}

				return false
			}

			p := parser.NewParser(file)
			p.ProcessPak(shouldProcess, func(entry *parser.PakEntrySet, _ *parser.PakFile) {
				results = append(results, entry)
			})
		}

		resultBytes := formatResults(results)
		err = ioutil.WriteFile(*output, resultBytes, 0644)

		if err != nil {
			panic(err)
		}
	},
}

func formatResults(result interface{}) []byte {
	var resultBytes []byte
	var err error

	if *format == "json" {
		if *pretty {
			resultBytes, err = json.MarshalIndent(result, "", "  ")
		} else {
			resultBytes, err = json.Marshal(result)
		}

		if err != nil {
			panic(err)
		}
	} else {
		panic("Unknown output format: " + *format)
	}

	return resultBytes
}
