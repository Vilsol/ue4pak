package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/MovieScene/Public/Channels/MovieSceneFloatChannel.h#L80
type FMovieSceneFloatValue struct {
	InValue float32 `json:"in_value"`
}

func (parser *PakParser) ReadFMovieSceneFloatValue() *FMovieSceneFloatValue {
	return &FMovieSceneFloatValue{
		InValue: parser.ReadFloat32(),
	}
}
