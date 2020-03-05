package parser

type FMovieSceneFrameRange struct {
	Value *TRange `json:"value"`
}

func (parser *PakParser) ReadFMovieSceneFrameRange() *FMovieSceneFrameRange {
	return &FMovieSceneFrameRange{
		Value: parser.ReadTRange("int32"),
	}
}
