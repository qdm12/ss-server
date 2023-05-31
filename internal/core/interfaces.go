package core

// SaltFilter is used to mitigate replay attacks by detecting repeated salts.
type SaltFilter interface {
	AddSalt(b []byte)
	IsSaltRepeated(b []byte) bool
}
