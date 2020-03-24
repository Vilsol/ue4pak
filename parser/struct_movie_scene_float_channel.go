package parser

// https://github.com/EpicGames/UnrealEngine/blob/4.22/Engine/Source/Runtime/MovieScene/Public/Channels/MovieSceneFloatChannel.h#L299
type FMovieSceneFloatChannel struct {
	PreInfinityExtrap  uint8                   `json:"pre_infinity_extrap"`
	PostInfinityExtrap uint8                   `json:"post_infinity_extrap"`
	Times              []FFrameNumber          `json:"times"`
	Values             []FMovieSceneFloatValue `json:"values"`
	DefaultValue       float32                 `json:"default_value"`
	HasDefaultValue    bool                    `json:"has_default_value"`
}

func (parser *PakParser) ReadFMovieSceneFloatChannel() *FMovieSceneFloatChannel {
	panic("ReadFMovieSceneFloatChannel is not implemented")
}
