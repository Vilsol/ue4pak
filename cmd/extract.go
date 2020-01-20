package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/Vilsol/ue4pak/parser"
	"github.com/fatih/color"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

var assets *[]string
var format *string
var output *string
var pretty *bool

type Result struct {
	Summary *parser.FPackageFileSummary `json:"summary"`
	Exports []ExportSet                 `json:"exports"`
}

type ExportSet struct {
	Export     *parser.FObjectExport  `json:"export"`
	Properties []*parser.FPropertyTag `json:"properties"`
}

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

		results := make([]*Result, 0)

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

				passed := false

				for _, pattern := range patterns {
					if pattern.Match(trimmed) {
						passed = true
						break
					}
				}

				if !passed {
					continue
				}

				if strings.HasSuffix(trimmed, "uasset") {
					fmt.Printf("Reading Record: %d: %#v\n", j, record)
					summaries[trimmed[0:strings.Index(trimmed, ".uasset")]] = record.ReadUAsset(file)
				}
			}

			// Second pass, parse exports
			for j, record := range pak.Index.Records {
				trimmed := trim(record.FileName)

				passed := false

				for _, pattern := range patterns {
					if pattern.Match(trimmed) {
						passed = true
						break
					}
				}

				if !passed {
					continue
				}

				if strings.HasSuffix(trimmed, "uexp") {
					summary, ok := summaries[trimmed[0:strings.Index(trimmed, ".uexp")]]

					if !ok {
						fmt.Printf("Unable to read record. Missing uasset: %d: %#v\n", j, record)
						continue
					}

					fmt.Printf("Reading Record: %d: %#v\n", j, record)

					exports := record.ReadUExp(file, summary)

					exportSet := make([]ExportSet, len(exports))

					i := 0
					for export, properties := range exports {
						exportSet[i] = ExportSet{
							Export:     export,
							Properties: properties,
						}
						i++
					}

					results = append(results, &Result{
						Summary: summary,
						Exports: exportSet,
					})
				}
			}
		}

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
