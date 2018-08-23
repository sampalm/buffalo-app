package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gobuffalo/buffalo/binding"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/pkg/errors"
)

type Post struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	Title     string       `json:"title" db:"title"`
	FileImage binding.File `db:"-" form:"FileImage"`
	FileName  string       `json:"file_name" db:"file_name"`
	Content   string       `json:"content" db:"content"`
	AuthorID  uuid.UUID    `json:"author_id" db:"author_id"`
}

type Posts []Post

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

//  Upload file to Disk and create a new post
func (p *Post) UploadAndCreate(tx *pop.Connection) (*validate.Errors, error) {
	// Check if a file was selected and is a valid file
	fExt := filepath.Ext(p.FileImage.Filename)
	exts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true}
	if _, ok := exts[fExt]; !ok || p.FileImage.Filename == "" {
		verrs := validate.NewErrors()
		verrs.Add("FileImage", "Invalid selected file")
		return verrs, nil
	}

	// Get files hashed name
	p.FileName = fmt.Sprint(GetMD5Hash(p.FileImage.Filename), fExt)

	// Creates uploads folder if it isnt exists yet
	dir := filepath.Join(".", "public", "uploads")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return validate.NewErrors(), errors.WithStack(err)
	}
	// Copy file to uploads folder
	f, err := os.Create(filepath.Join(dir, p.FileName))
	if err != nil {
		return validate.NewErrors(), errors.WithStack(err)
	}
	defer f.Close()
	if _, err = io.Copy(f, p.FileImage); err != nil {
		return validate.NewErrors(), errors.WithStack(err)
	}
	// Save post into the DB
	return tx.ValidateAndCreate(p)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidadeAndCreate, pop.ValidateAndUpdate) method.
func (p *Post) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: p.Title, Name: "Title"},
		&validators.StringIsPresent{Field: p.Content, Name: "Content"},
	), nil
}
