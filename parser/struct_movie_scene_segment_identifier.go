package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/MovieScene/Public/Evaluation/MovieSceneSegment.h#L31
type FMovieSceneSegmentIdentifier struct {
	IdentifierIndex int32 `json:"identifier_index"`
}

func (parser *PakParser) ReadFMovieSceneSegmentIdentifier() *FMovieSceneSegmentIdentifier {
	return &FMovieSceneSegmentIdentifier{
		IdentifierIndex: parser.ReadInt32(),
	}
}
