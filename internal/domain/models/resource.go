package models

type ResourceType int

const (
	ScriptType = iota
	CSSType
	ImageType
)

// Resource is a structure that can store `CSS`, `JS Scripts`, `Images`.
// All of them are distinguished by the `Type` field.
type Resource struct {
	ID      int          `db:"id"`
	Name    string       `db:"name"`
	Content []byte       `db:"content"`
	Type    ResourceType `db:"type"`
}
