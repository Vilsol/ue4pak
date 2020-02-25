package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var PakFile string
var LogLevel string
var ForceColors bool
var NoPreload bool

var rootCmd = &cobra.Command{
	Use:   "ue4pak",
	Short: "ue4pak parses and extracts data from UE4 Pak files",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		level, err := log.ParseLevel(LogLevel)

		if err != nil {
			panic(err)
		}

		log.SetFormatter(&log.TextFormatter{
			ForceColors: ForceColors,
		})
		log.SetOutput(os.Stdout)
		log.SetLevel(level)

		viper.Set("NoPreload", NoPreload)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&PakFile, "pak", "p", "", "The path to pak file (supports glob) (required)")
	rootCmd.PersistentFlags().StringVar(&LogLevel, "log", "info", "The log level to output")
	rootCmd.PersistentFlags().BoolVar(&ForceColors, "colors", false, "Force output with colors")
	rootCmd.PersistentFlags().BoolVar(&NoPreload, "no-preload", false, "Do not preload data (slower, but guaranteed to read)")
	rootCmd.MarkPersistentFlagRequired("pak")
}
