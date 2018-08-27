package models

import (
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/pkg/errors"
)

type Tag struct {
	ID   uuid.UUID `json:"id" db:"id"`
	Name string    `json:"name" db:"name"`
}

type Tags []Tag

// Create a new post tag
func (t *Tag) Generate(tx *pop.Connection) error {
	q := tx.Where("Name = ?", t.Name)
	exists, err := q.Exists(t)
	if err != nil {
		return errors.WithStack(err)
	}
	if !exists {
		return tx.Create(t)
	}
	if err := q.First(t); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidadeAndCreate, pop.ValidateAndUpdate) method.
func (t *Tag) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: t.Name, Name: "Name"},
	), nil
}
