package parser

type FMovieSceneFloatValue struct {
	InValue float32 `json:"in_value"`
}

func (parser *PakParser) ReadFMovieSceneFloatValue() *FMovieSceneFloatValue {
	return &FMovieSceneFloatValue{
		InValue: parser.ReadFloat32(),
	}
}
