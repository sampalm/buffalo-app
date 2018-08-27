package models

import (
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
)

type TagPost struct {
	ID     uuid.UUID `json:"id" db:"id"`
	PostID uuid.UUID `json:"post_id" db:"post_id"`
	TagID  uuid.UUID `json:"tag_id" db:"tag_id"`
}

type TagsPosts []TagPost

// TableName overrides the table name used by Pop.
func (t *TagPost) TableName() string {
	return "tags_posts"
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidadeAndCreate, pop.ValidateAndUpdate) method.
func (t *TagPost) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
