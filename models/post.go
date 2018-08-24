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

// Generate a MD5 Hash from a string
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Copy a file to disk
func CopyToDisk(file binding.File, filename, ext string) error {
	// Creates uploads folder if it isnt exists yet
	dir := filepath.Join(".", "public", "uploads")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// Copy file to uploads folder
	f, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = io.Copy(f, file); err != nil {
		return err
	}
	return nil
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

	// Try to copy the file
	if err := CopyToDisk(p.FileImage, p.FileName, fExt); err != nil {
		return validate.NewErrors(), errors.WithStack(err)
	}

	// Save post into the DB
	return tx.ValidateAndCreate(p)
}

//  Upload file to Disk and update the users post
func (p *Post) UploadAndUpdated(tx *pop.Connection) (*validate.Errors, error) {
	// Check if a file was selected
	fExt := filepath.Ext(p.FileImage.Filename)
	if p.FileImage.Filename != "" {
		exts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true}
		// Check if the file extension is valid
		if _, ok := exts[fExt]; !ok {
			verrs := validate.NewErrors()
			verrs.Add("FileImage", "Invalid selected file")
			return verrs, nil
		}
		// Get files hashed name
		p.FileName = fmt.Sprint(GetMD5Hash(p.FileImage.Filename), fExt)
		// Try to copy the file
		if err := CopyToDisk(p.FileImage, p.FileName, fExt); err != nil {
			return validate.NewErrors(), errors.WithStack(err)
		}
	}

	// Just update the users post without change the file image
	return tx.ValidateAndUpdate(p)
}

//  Upload file to Disk and create a new post
func (p *Post) DeleteFile(tx *pop.Connection) error {
	// Query all posts that have that filename
	ct, err := tx.Where("file_name = ?", p.FileName).Count(Post{})
	if err != nil {
		return errors.WithStack(err)
	}
	// If there are only a single post using that image it can be removed from disk
	if ct == 0 {
		dir := filepath.Join(".", "public", "uploads", p.FileName)
		if err = os.Remove(dir); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidadeAndCreate, pop.ValidateAndUpdate) method.
func (p *Post) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: p.Title, Name: "Title"},
		&validators.StringIsPresent{Field: p.Content, Name: "Content"},
	), nil
}
