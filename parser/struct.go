package parser

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

type StructType struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func (parser *PakParser) ReadStruct(property *StructProperty, size int32, depth int) (interface{}, bool) {
	switch strings.Trim(property.Type, "\x00") {
	case "Vector":
		return parser.ReadFVector(), true
	case "LinearColor":
		return parser.ReadFLinearColor(), true
	case "Vector2D":
		return parser.ReadFVector2D(), true
	case "IntPoint":
		return parser.ReadFIntPoint(), true
	case "Rotator":
		return parser.ReadFRotator(), true
	case "Quat":
		return parser.ReadFQuat(), true
	case "Vector4":
		return parser.ReadFVector4(), true
	case "Color":
		return parser.ReadFColor(), true
	case "Box":
		return parser.ReadFBox(), true
	case "FrameNumber":
		return parser.ReadFFrameNumber(), true
	case "MovieSceneSequenceID":
		return parser.ReadFMovieSceneSequenceID(), true
	case "Box2D":
		return parser.ReadFBox2D(), true
	case "MovieSceneTrackIdentifier":
		return parser.ReadFMovieSceneTrackIdentifier(), true
	case "MovieSceneEvaluationKey":
		return parser.ReadFMovieSceneEvaluationKey(), true
	case "MovieSceneSegmentIdentifier":
		return parser.ReadFMovieSceneSegmentIdentifier(), true
	case "MovieSceneFloatValue":
		return parser.ReadFMovieSceneFloatValue(), true
	case "RichCurveKey":
		return parser.ReadFRichCurveKey(), true
	case "MovieSceneFrameRange":
		return parser.ReadFMovieSceneFrameRange(), true
	case "Guid":
		if size == 16 {
			return parser.ReadFGuid(), true
		}

		// TODO Something is not right
		// return parser.ReadFGuid()
		fallthrough
	case "VectorMaterialInput":
		fallthrough
	case "ExpressionInput":
		fallthrough
	case "ScalarMaterialInput":
		fallthrough
	case "ColorMaterialInput":
		fallthrough
	case "PerPlatformFloat":
		fallthrough
	case "SkeletalMeshSamplingLODBuiltData":
		fallthrough
	case "PointerToUberGraphFrame":
		fallthrough
	case "FontData":
		fallthrough
	case "FontCharacter":
		fallthrough
	case "MaterialAttributesInput":
		fallthrough
	case "MovieSceneByteChannel":
		fallthrough
	case "MovieSceneEventParameters":
		fallthrough
	case "SoftClassPath":
		fallthrough
	case "MovieSceneParticleChannel":
		fallthrough
	case "InventoryItem":
		fallthrough
	case "SmartName":
		fallthrough
	case "PerPlatformInt":
		fallthrough
	case "MovieSceneSegment":
		fallthrough
	case "SectionEvaluationDataTree":
		fallthrough
	case "MovieSceneEvalTemplatePtr":
		fallthrough
	case "MovieSceneTrackImplementationPtr":
		fallthrough
	case "MovieSceneEvaluationTrack":
		fallthrough
	case "LevelSequenceBindingReferenceArray":
		// TODO Read types correctly
		log.Debugf("%sUnread StructProperty Type [%d]: %s", d(depth), size, strings.Trim(property.Type, "\x00"))
		// fmt.Println(utils.HexDump(data[offset:]))
		if size > 0 {
			parser.Read(size)
		}
		return nil, true
	default:
		// All others are fine
		break
	}

	return nil, false
}
