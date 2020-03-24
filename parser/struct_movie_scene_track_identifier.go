package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/MovieScene/Public/Evaluation/MovieSceneTrackIdentifier.h#L11
type FMovieSceneTrackIdentifier struct {
	Value uint32 `json:"value"`
}

func (parser *PakParser) ReadFMovieSceneTrackIdentifier() *FMovieSceneTrackIdentifier {
	return &FMovieSceneTrackIdentifier{
		Value: parser.ReadUint32(),
	}
}
