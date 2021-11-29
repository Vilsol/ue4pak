package parser

import (
	"context"
	"runtime/debug"
	"strings"

	"github.com/rs/zerolog/log"
)

func (parser *PakParser) ProcessPak(ctx context.Context, parseFile func(string) bool, handleEntry func(string, *PakEntrySet, *PakFile)) {
	pak := parser.Parse(ctx)

	if pak == nil {
		return
	}

	summaries := make(map[string]*FPackageFileSummary, 0)

	// First pass, parse summaries
	for j, record := range pak.Index.Records {
		trimmed := strings.Trim(record.FileName, "\x00")

		if parseFile != nil {
			if !parseFile(trimmed) {
				continue
			}
		}

		if strings.HasSuffix(trimmed, "uasset") {
			offset := record.FileOffset + int64(pak.Footer.HeaderSize())
			log.Ctx(ctx).Info().Msgf("Reading Summary: %d [%x-%x]: %s", j, offset, offset+record.FileSize, trimmed)
			summaries[trimmed[0:strings.Index(trimmed, ".uasset")]] = record.ReadUAsset(pak, parser)
			summaries[trimmed[0:strings.Index(trimmed, ".uasset")]].Record = record
		}
	}

	// Second pass, parse exports
	for j, record := range pak.Index.Records {
		trimmed := strings.Trim(record.FileName, "\x00")

		if parseFile != nil {
			if !parseFile(trimmed) {
				continue
			}
		}

		if strings.HasSuffix(trimmed, "uexp") {
			summary, ok := summaries[trimmed[0:strings.Index(trimmed, ".uexp")]]

			offset := record.FileOffset + int64(pak.Footer.HeaderSize())

			if !ok {
				log.Ctx(ctx).Error().Msgf("Unable to read record. Missing uasset: %d [%x-%x]: %s", j, offset, offset+record.FileSize, trimmed)
				continue
			}

			log.Ctx(ctx).Info().Msgf("Reading Record: %d [%x-%x]: %s", j, offset, offset+record.FileSize, trimmed)

			output := make(chan map[*FObjectExport]*ExportData)

			go func() {
				defer func() {
					if err := recover(); err != nil {
						log.Ctx(ctx).Error().Str("stack", string(debug.Stack())).Msgf("error parsing record: %v", err)
						output <- make(map[*FObjectExport]*ExportData)
					}
				}()
				output <- record.ReadUExp(ctx, pak, parser, summary)
			}()

			exports := <-output

			exportSet := make([]PakExportSet, len(exports))

			i := 0
			for export, data := range exports {
				exportSet[i] = PakExportSet{
					Export: export,
					Data:   data,
				}
				i++
			}

			if handleEntry != nil {
				handleEntry(trimmed, &PakEntrySet{
					ExportRecord: record,
					Summary:      summary,
					Exports:      exportSet,
				}, pak)
			}
		}
	}
}
