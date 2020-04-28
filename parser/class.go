package parser

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

type ClassResolver func(parser *PakParser, export *FObjectExport, size int32, uAsset *FPackageFileSummary) interface{}

var classResolvers = map[string]ClassResolver{
	"DataTable": func(parser *PakParser, export *FObjectExport, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadUDataTable(uAsset)
	},
	"ObjectProperty": func(parser *PakParser, export *FObjectExport, size int32, uAsset *FPackageFileSummary) interface{} {
		// TODO Figure out
		parser.Read(24)
		return parser.ReadFPackageIndex(uAsset.Imports, uAsset.Exports)
	},
	"BoolProperty": func(parser *PakParser, export *FObjectExport, size int32, uAsset *FPackageFileSummary) interface{} {
		// TODO Figure out
		parser.Read(25)
		return parser.Read(1)[0] != 0
	},
	"StructProperty": func(parser *PakParser, export *FObjectExport, size int32, uAsset *FPackageFileSummary) interface{} {
		// TODO Figure out
		parser.Read(24)
		return parser.ReadFPackageIndex(uAsset.Imports, uAsset.Exports)
	},
	"DelegateProperty": func(parser *PakParser, export *FObjectExport, size int32, uAsset *FPackageFileSummary) interface{} {
		// TODO Figure out
		parser.Read(24)
		return parser.ReadFPackageIndex(uAsset.Imports, uAsset.Exports)
	},
}

type ClassType struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func (parser *PakParser) ReadClass(export *FObjectExport, size int32, uAsset *FPackageFileSummary) (interface{}, bool) {
	var className string

	if classNameTemp := export.TemplateIndex.ClassName(); classNameTemp != nil {
		className = *classNameTemp
	} else {
		return nil, false
	}

	trimmedType := strings.Trim(className, "\x00")

	resolver, ok := classResolvers[trimmedType]

	if !ok {
		return nil, false
	}

	if resolver != nil {
		value := resolver(parser, export, size, uAsset)

		if value != nil {
			return value, true
		}
	}

	// TODO Read types correctly
	log.Warningf("Unread Class Type [%d]: %s", size, trimmedType)
	// fmt.Println(utils.HexDump(data[offset:]))
	if size > 0 {
		parser.Read(size)
	}

	return nil, true
}

func RegisterClassResolver(classType string, resolver ClassResolver) {
	classResolvers[classType] = resolver
}
