package model

// Tag is a free-form label: a category says what kind of spend it is,
// a tag says which trip or project it belongs to.
type Tag struct {
	ID   int64
	Name string
}
