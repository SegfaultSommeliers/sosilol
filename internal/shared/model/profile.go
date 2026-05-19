package model

type Profile struct {
	ID        int64
	Login     string
	AvatarURL string
	Pastes    []Paste
}
