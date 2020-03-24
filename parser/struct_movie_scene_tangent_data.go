package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/MovieScene/Public/Channels/MovieSceneFloatChannel.h#L17
type FMovieSceneTangentData struct {
	ArriveTangent       float32 `json:"arrive_tangent"`
	LeaveTangent        float32 `json:"leave_tangent"`
	TangentWeightMode   uint8   `json:"tangent_weight_mode"`
	ArriveTangentWeight float32 `json:"arrive_tangent_weight"`
	LeaveTangentWeight  float32 `json:"leave_tangent_weight"`
}

func (parser *PakParser) ReadFMovieSceneTangentData() *FMovieSceneTangentData {
	return &FMovieSceneTangentData{
		ArriveTangent:       parser.ReadFloat32(),
		LeaveTangent:        parser.ReadFloat32(),
		TangentWeightMode:   parser.Read(1)[0],
		ArriveTangentWeight: parser.ReadFloat32(),
		LeaveTangentWeight:  parser.ReadFloat32(),
	}
}
