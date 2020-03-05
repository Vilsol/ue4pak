package parser

type FMovieSceneSegmentIdentifier struct {
	IdentifierIndex int32 `json:"identifier_index"`
}

func (parser *PakParser) ReadFMovieSceneSegmentIdentifier() *FMovieSceneSegmentIdentifier {
	return &FMovieSceneSegmentIdentifier{
		IdentifierIndex: parser.ReadInt32(),
	}
}
