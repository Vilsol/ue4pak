package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/MovieScene/Public/MovieSceneSequenceID.h#L11
type FMovieSceneSequenceID struct {
	InValue uint32 `json:"in_value"`
}

func (parser *PakParser) ReadFMovieSceneSequenceID() *FMovieSceneSequenceID {
	return &FMovieSceneSequenceID{
		InValue: parser.ReadUint32(),
	}
}
