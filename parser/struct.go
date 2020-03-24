package parser

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

type StructResolver func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{}

var structResolvers = map[string]StructResolver{
	"Vector": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFVector()
	},
	"LinearColor": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFLinearColor()
	},
	"Vector2D": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFVector2D()
	},
	"IntPoint": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFIntPoint()
	},
	"Rotator": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFRotator()
	},
	"Quat": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFQuat()
	},
	"Vector4": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFVector4()
	},
	"Color": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFColor()
	},
	"Box": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFBox()
	},
	"FrameNumber": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFFrameNumber()
	},
	"MovieSceneSequenceID": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFMovieSceneSequenceID()
	},
	"Box2D": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFBox2D()
	},
	"MovieSceneTrackIdentifier": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFMovieSceneTrackIdentifier()
	},
	"MovieSceneEvaluationKey": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFMovieSceneEvaluationKey()
	},
	"MovieSceneSegmentIdentifier": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFMovieSceneSegmentIdentifier()
	},
	"MovieSceneFloatValue": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFMovieSceneFloatValue()
	},
	"RichCurveKey": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFRichCurveKey()
	},
	"MovieSceneFrameRange": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		return parser.ReadFMovieSceneFrameRange()
	},
	"Guid": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		if size == 16 {
			return parser.ReadFGuid()
		}

		// TODO Something is not right
		// return parser.ReadFGuid()
		return nil
	},
	"VectorMaterialInput": nil,
	"ExpressionInput": func(parser *PakParser, property *StructProperty, size int32, uAsset *FPackageFileSummary) interface{} {
		// Unsure if read correctly
		if size != 40 {
			return nil
		}

		return parser.ReadFExpressionInput(uAsset.Names)
	},
	"ScalarMaterialInput":                nil,
	"ColorMaterialInput":                 nil,
	"PerPlatformFloat":                   nil,
	"SkeletalMeshSamplingLODBuiltData":   nil,
	"PointerToUberGraphFrame":            nil,
	"FontData":                           nil,
	"FontCharacter":                      nil,
	"MaterialAttributesInput":            nil,
	"MovieSceneByteChannel":              nil,
	"MovieSceneEventParameters":          nil,
	"SoftClassPath":                      nil,
	"MovieSceneParticleChannel":          nil,
	"SmartName":                          nil,
	"PerPlatformInt":                     nil,
	"MovieSceneSegment":                  nil,
	"SectionEvaluationDataTree":          nil,
	"MovieSceneEvalTemplatePtr":          nil,
	"MovieSceneTrackImplementationPtr":   nil,
	"MovieSceneEvaluationTrack":          nil,
	"MovieSceneFloatChannel":             nil,
	"LevelSequenceBindingReferenceArray": nil,
}

type StructType struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func (parser *PakParser) ReadStruct(property *StructProperty, size int32, uAsset *FPackageFileSummary, depth int) (interface{}, bool) {
	trimmedType := strings.Trim(property.Type, "\x00")

	resolver, ok := structResolvers[trimmedType]

	if !ok {
		return nil, false
	}

	if resolver != nil {
		value := resolver(parser, property, size, uAsset)

		if value != nil {
			return value, true
		}
	}

	// TODO Read types correctly
	log.Warningf("%sUnread StructProperty Type [%d]: %s", d(depth), size, trimmedType)
	// fmt.Println(utils.HexDump(data[offset:]))
	if size > 0 {
		parser.Read(size)
	}

	return nil, true
}

func RegisterStructResolver(structType string, resolver StructResolver) {
	structResolvers[structType] = resolver
}
