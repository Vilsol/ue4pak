package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/MovieScene/Public/Channels/MovieSceneFloatChannel.h#L80
type FMovieSceneFloatValue struct {
	Value       float32                 `json:"value"`
	InterpMode  uint8                   `json:"interp_mode"`
	TangentMode uint8                   `json:"tangent_mode"`
	Tangent     *FMovieSceneTangentData `json:"tangent"`
}

func (parser *PakParser) ReadFMovieSceneFloatValue() *FMovieSceneFloatValue {
	return &FMovieSceneFloatValue{
		Value:       parser.ReadFloat32(),
		InterpMode:  parser.Read(1)[0],
		TangentMode: parser.Read(1)[0],
		Tangent:     parser.ReadFMovieSceneTangentData(),
	}
}
