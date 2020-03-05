package parser

type FMovieSceneEvaluationKey struct {
	SequenceID      *FMovieSceneSequenceID      `json:"sequence_id"`
	TrackIdentifier *FMovieSceneTrackIdentifier `json:"track_identifier"`
	SectionIndex    uint32                      `json:"section_index"`
}

func (parser *PakParser) ReadFMovieSceneEvaluationKey() *FMovieSceneEvaluationKey {
	return &FMovieSceneEvaluationKey{
		SequenceID:      parser.ReadFMovieSceneSequenceID(),
		TrackIdentifier: parser.ReadFMovieSceneTrackIdentifier(),
		SectionIndex:    parser.ReadUint32(),
	}
}
