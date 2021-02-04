package parser

import (
	"fmt"
	"runtime/debug"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (parser *PakParser) ProcessPak(parseFile func(string) bool, handleEntry func(string, *PakEntrySet, *PakFile)) {
	pak := parser.Parse()

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
			offset := record.FileOffset + pak.Footer.HeaderSize()
			log.Infof("Reading Summary: %d [%x-%x]: %s\n", j, offset, offset+record.FileSize, trimmed)
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

			offset := record.FileOffset + pak.Footer.HeaderSize()

			if !ok {
				log.Errorf("Unable to read record. Missing uasset: %d [%x-%x]: %s\n", j, offset, offset+record.FileSize, trimmed)
				continue
			}

			log.Infof("Reading Record: %d [%x-%x]: %s\n", j, offset, offset+record.FileSize, trimmed)

			output := make(chan map[*FObjectExport]*ExportData)

			go func() {
				defer func() {
					if err := recover(); err != nil {
						log.Errorf("error parsing record: %v", err)
						fmt.Println(string(debug.Stack()))
						output <- make(map[*FObjectExport]*ExportData)
					}
				}()
				output <- record.ReadUExp(pak, parser, summary)
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
