package main

import (
	"context"
	"fmt"
	"github.com/Vilsol/ue4pak/parser"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAllAsFiles(t *testing.T) {
	color.NoColor = false

	paks, err := filepath.Glob("paks/*.pak")

	if err != nil {
		panic(err)
	}

	for _, f := range paks {
		log.Info().Msgf("Parsing file: %s", f)

		file, err := os.OpenFile(f, os.O_RDONLY, 0644)

		if err != nil {
			panic(err)
		}

		p := parser.NewParser(file)
		pak := p.Parse(context.Background())

		summaries := make(map[string]*parser.FPackageFileSummary, 0)

		// First pass, parse summaries
		for j, record := range pak.Index.Records {
			trimmed := strings.Trim(record.FileName, "\x00")
			if strings.HasSuffix(trimmed, "uasset") {
				fmt.Printf("Reading Record: %d: %#v\n", j, record)
				summaries[trimmed[0:strings.Index(trimmed, ".uasset")]] = record.ReadUAsset(pak, p)
			}
		}

		// Second pass, parse exports
		for j, record := range pak.Index.Records {
			trimmed := strings.Trim(record.FileName, "\x00")
			if strings.HasSuffix(trimmed, "uexp") {
				summary, ok := summaries[trimmed[0:strings.Index(trimmed, ".uexp")]]

				if !ok {
					fmt.Printf("Unable to read record. Missing uasset: %d: %#v\n", j, record)
					continue
				}

				fmt.Printf("Reading Record: %d: %#v\n", j, record)

				record.ReadUExp(context.Background(), pak, p, summary)
			}
		}
	}
}

func TestParseAllAsBytes(t *testing.T) {
	color.NoColor = false

	paks, err := filepath.Glob("paks/*.pak")

	if err != nil {
		panic(err)
	}

	for _, f := range paks {
		log.Info().Msgf("Parsing file: %s", f)

		data, err := ioutil.ReadFile(f)

		if err != nil {
			panic(err)
		}

		reader := &parser.PakByteReader{
			Bytes: data,
		}

		p := parser.NewParser(reader)
		pak := p.Parse(context.Background())

		summaries := make(map[string]*parser.FPackageFileSummary, 0)

		// First pass, parse summaries
		for j, record := range pak.Index.Records {
			trimmed := strings.Trim(record.FileName, "\x00")
			if strings.HasSuffix(trimmed, "uasset") {
				fmt.Printf("Reading Record: %d: %#v\n", j, record)
				summaries[trimmed[0:strings.Index(trimmed, ".uasset")]] = record.ReadUAsset(pak, p)
			}
		}

		// Second pass, parse exports
		for j, record := range pak.Index.Records {
			trimmed := strings.Trim(record.FileName, "\x00")
			if strings.HasSuffix(trimmed, "uexp") {
				summary, ok := summaries[trimmed[0:strings.Index(trimmed, ".uexp")]]

				if !ok {
					fmt.Printf("Unable to read record. Missing uasset: %d: %#v\n", j, record)
					continue
				}

				fmt.Printf("Reading Record: %d: %#v\n", j, record)

				record.ReadUExp(context.Background(), pak, p, summary)
			}
		}
	}
}
