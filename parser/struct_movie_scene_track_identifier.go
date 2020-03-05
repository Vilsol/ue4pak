package parser

type FMovieSceneTrackIdentifier struct {
	Value uint32 `json:"value"`
}

func (parser *PakParser) ReadFMovieSceneTrackIdentifier() *FMovieSceneTrackIdentifier {
	return &FMovieSceneTrackIdentifier{
		Value: parser.ReadUint32(),
	}
}
