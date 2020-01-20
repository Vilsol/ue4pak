package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var PakFile string

var rootCmd = &cobra.Command{
	Use:   "ue4pak",
	Short: "ue4pak parses and extracts data from UE4 Pak files",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&PakFile, "pak", "p", "", "The path to pak file (supports glob) (required)")
	rootCmd.MarkPersistentFlagRequired("pak")
}
