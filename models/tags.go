package models

import (
	"regexp"
	"strings"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/pkg/errors"
)

type Tag struct {
	ID   uuid.UUID `json:"id" db:"id"`
	Name string    `json:"name" db:"name"`
	Code int       `json:"code" db:"code" rw:"r"`
}

type Tags []Tag

type StringIsRegular struct {
	Field   string
	Name    string
	Message string
	tx      *pop.Connection
}

// Generate a new categorize tag
func (t *Tag) Generate(tx *pop.Connection) (*validate.Errors, error) {
	// Make refex to say we only want
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return validate.NewErrors(), err
	}
	t.Name = strings.ToLower(reg.ReplaceAllString(t.Name, ""))

	// Check if the tag already exists
	q := tx.Where("Name = ?", t.Name)
	exists, err := q.Exists(t)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if exists {
		verrs := validate.NewErrors()
		verrs.Add("Name", "Tag Name is already being used.")
		return verrs, nil
	}

	// Create a new tag
	return tx.ValidateAndCreate(t)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidadeAndCreate, pop.ValidateAndUpdate) method.
func (t *Tag) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: t.Name, Name: "Name"},
	), nil
}
