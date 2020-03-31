package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Vilsol/ue4pak/parser"
	"github.com/fatih/color"

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

			p := parser.NewParser(file)
			p.ProcessPak(nil, nil)
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
