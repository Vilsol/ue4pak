package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/MovieScene/Public/MovieSceneFrameMigration.h#L16
type FMovieSceneFrameRange struct {
	Value *TRange `json:"value"`
}

func (parser *PakParser) ReadFMovieSceneFrameRange() *FMovieSceneFrameRange {
	return &FMovieSceneFrameRange{
		Value: parser.ReadTRange("int32"),
	}
}
