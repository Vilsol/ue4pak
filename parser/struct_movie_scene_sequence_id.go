package parser

type FMovieSceneSequenceID struct {
	InValue uint32 `json:"in_value"`
}

func (parser *PakParser) ReadFMovieSceneSequenceID() *FMovieSceneSequenceID {
	return &FMovieSceneSequenceID{
		InValue: parser.ReadUint32(),
	}
}
