package cmd

import (
	"fmt"
	"github.com/gobwas/glob"
	"os"
	"path/filepath"
	"strings"

	"github.com/Vilsol/ue4pak/parser"
	"github.com/fatih/color"

	"github.com/spf13/cobra"
)

var testAssets *[]string

func init() {
	testAssets = testCmd.Flags().StringSliceP("assets", "a", []string{}, "Comma-separated list of asset paths to extract. (supports glob)")

	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test parse the provided paks",
	Run: func(cmd *cobra.Command, args []string) {
		color.NoColor = false
		paks, err := filepath.Glob(cmd.Flag("pak").Value.String())

		patterns := make([]glob.Glob, len(*testAssets))
		for i, asset := range *testAssets {
			patterns[i] = glob.MustCompile(asset)
		}

		if err != nil {
			panic(err)
		}

		fmt.Println(paks)

		for _, f := range paks {
			fmt.Println("Parsing file:", f)

			file, err := os.OpenFile(f, os.O_RDONLY, 0644)

			if err != nil {
				panic(err)
			}

			shouldProcess := func(name string) bool {
				if len(patterns) == 0 {
					return true
				}

				for _, pattern := range patterns {
					if pattern.Match(name) {
						return true
					}
				}

				return false
			}

			p := parser.NewParser(file)
			p.ProcessPak(shouldProcess, nil)
			/*
				f, err := os.OpenFile("dump.txt", os.O_WRONLY | os.O_CREATE, 0644)
				fmt.Println(err)
				spew.Fdump(f, pak)
				f.Close()
				return
			*/
		}
	},
}

func trim(s string) string {
	return strings.Trim(s, "\x00")
}
