package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Core/Public/Math/Rotator.h#L18
type FRotator struct {
	Pitch float32 `json:"pitch"`
	Yaw   float32 `json:"yaw"`
	Roll  float32 `json:"roll"`
}

func (parser *PakParser) ReadFRotator() *FRotator {
	return &FRotator{
		Pitch: parser.ReadFloat32(),
		Yaw:   parser.ReadFloat32(),
		Roll:  parser.ReadFloat32(),
	}
}
