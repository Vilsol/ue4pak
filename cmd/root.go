package cmd

import (
	"github.com/rs/zerolog"
	"os"
	"time"

	_ "github.com/Vilsol/ue4pak/parser/games/satisfactory"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var PakFile string
var LogLevel string
var ForceColors bool
var NoPreload bool

var rootCmd = &cobra.Command{
	Use:   "ue4pak",
	Short: "ue4pak parses and extracts data from UE4 Pak files",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		level, err := zerolog.ParseLevel(LogLevel)
		if err != nil {
			log.Err(err).Msg("Invalid log level")
		}

		zerolog.SetGlobalLevel(level)

		log.Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Logger()

		viper.Set("NoPreload", NoPreload)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Msg(err.Error())
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
