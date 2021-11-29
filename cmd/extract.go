package cmd

import (
	"encoding/json"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Vilsol/ue4pak/parser"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"

	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

var assets *[]string

func init() {
	assets = extractCmd.Flags().StringSliceP("assets", "a", []string{}, "Comma-separated list of asset paths to extract. (supports glob) (required)")
	format = extractCmd.Flags().StringP("format", "f", "json", "Output format type")
	output = extractCmd.Flags().StringP("output", "o", "extracted.json", "Output file (or directory if --split)")
	split = extractCmd.Flags().Bool("split", false, "Whether output should be split into a file per asset")
	pretty = extractCmd.Flags().Bool("pretty", false, "Whether to output in a pretty format")

	extractCmd.Flags().Bool("with-index", false, "Whether to output FPackageIndex")
	extractCmd.Flags().Bool("with-names", false, "Whether to output names")

	_ = viper.BindPFlag("with-index", extractCmd.Flags().Lookup("with-index"))
	_ = viper.BindPFlag("with-names", extractCmd.Flags().Lookup("with-names"))

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
			log.Info().Msgf("Parsing file: %s", f)

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

			ctx := log.Logger.WithContext(cmd.Context())

			p := parser.NewParser(file)
			p.ProcessPak(ctx, shouldProcess, func(name string, entry *parser.PakEntrySet, _ *parser.PakFile) {
				if *split {
					destination := filepath.Join(*output, name+"."+*format)
					err := os.MkdirAll(filepath.Dir(destination), 0755)
					if err != nil {
						panic(err)
					}

					log.Info().Msgf("Writing Result: %s", destination)
					resultBytes := formatResults(entry)
					err = ioutil.WriteFile(destination, resultBytes, 0644)
					if err != nil {
						panic(err)
					}
				} else {
					results = append(results, entry)
				}
			})
		}

		/*
			if c, ok := x.Reference.(*FObjectExport); ok {
				if strings.Trim(c.ObjectName, "\x00") == "BPD_ResearchTreeNode_C" {
					uAsset.ParseObject(parser, c, pak, record)
				}
			}
		*/

		if !*split {
			resultBytes := formatResults(results)
			err = ioutil.WriteFile(*output, resultBytes, 0644)
		}

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
