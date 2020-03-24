package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Engine/Classes/Curves/RichCurve.h#L74
type FRichCurveKey struct {
	InterpMode          uint8   `json:"interp_mode"`
	TangentMode         uint8   `json:"tangent_mode"`
	TangentWeightMode   uint8   `json:"tangent_weight_mode"`
	Time                float32 `json:"time"`
	ArriveTangent       float32 `json:"arrive_tangent"`
	ArriveTangentWeight float32 `json:"arrive_tangent_weight"`
	LeaveTangent        float32 `json:"leave_tangent"`
	LeaveTangentWeight  float32 `json:"leave_tangent_weight"`
}

func (parser *PakParser) ReadFRichCurveKey() *FRichCurveKey {
	return &FRichCurveKey{
		InterpMode:          parser.Read(1)[0],
		TangentMode:         parser.Read(1)[0],
		TangentWeightMode:   parser.Read(1)[0],
		Time:                parser.ReadFloat32(),
		ArriveTangent:       parser.ReadFloat32(),
		ArriveTangentWeight: parser.ReadFloat32(),
		LeaveTangent:        parser.ReadFloat32(),
		LeaveTangentWeight:  parser.ReadFloat32(),
	}
}
