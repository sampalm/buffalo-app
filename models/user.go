package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID              uuid.UUID `json:"id" db:"id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	Name            string    `json:"name" db:"name"`
	Username        string    `json:"username" db:"username"`
	Email           string    `json:"email" db:"email"`
	Admin           bool      `json:"admin" db:"admin"`
	PasswordHash    string    `json:"-" db:"password_hash"`
	Password        string    `json:"-" db:"-"`
	PasswordConfirm string    `json:"-" db:"-"`
	Provider        string    `json:"provider" db:"provider"`
	ProviderID      string    `json:"provider_id" db:"provider_id"`
}

type ItsAvailable struct {
	Name   string
	Field  string
	Field2 string
	tx     *pop.Connection
}

// Create a new user into database
func (u User) Create(tx *pop.Connection) (*validate.Errors, error) {
	// Default user account
	u.Admin = false
	// Validade user password
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return validate.NewErrors(), errors.WithStack(err)
	}
	u.PasswordHash = string(pwdHash)
	return tx.ValidateAndCreate(&u)
}

func (u User) Update(tx *pop.Connection) (*validate.Errors, error) {
	// Validate user password
	if u.Password != "" {
		if check := validate.Validate(
			&validators.StringLengthInRange{Field: u.Password, Name: "Password", Min: 6, Max: 20, Message: "Password is too weak."},
			&validators.StringsMatch{Field: u.Password, Field2: u.PasswordConfirm, Name: "PasswordConfirm", Message: "Passwords do not match"},
		); check != nil {
			return check, nil
		}
		pwdHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return validate.NewErrors(), errors.WithStack(err)
		}
		u.PasswordHash = string(pwdHash)
	}
	// Validate user email
	if u.Email != "" {
		if check := validate.Validate(
			&validators.EmailLike{Field: u.Email, Name: "Email"},
		); check != nil {
			return check, nil
		}
		exists, err := tx.Where("email = ? AND id != ?", u.Email, u.ID).Exists(u)
		if err != nil {
			return validate.NewErrors(), errors.WithStack(err)
		}
		if exists {
			verrs := validate.NewErrors()
			verrs.Add("Email", "Email is already being used.")
			return verrs, nil
		}
	}

	return validate.NewErrors(), tx.Update(&u)
}

func (u *User) Authorize(tx *pop.Connection) error {
	// Check if email is into DB
	err := tx.Where("email = ?", u.Email).First(u)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return errors.New(fmt.Sprintf("Email %s not found", u.Email))
		}
		return errors.WithStack(err)
	}
	// Confirm that the given password matches
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(u.Password))
	if err != nil {
		return errors.New("Invalid password")
	}
	return nil
}

func (u *User) OAuthAndSave(tx *pop.Connection) error {
	exists, err := tx.Where("username = ?", u.Username).Exists(u)
	if err != nil {
		return errors.WithStack(err)
	}
	if exists {
		u.Username = fmt.Sprintf("%s%d", u.Username, u.CreatedAt.UnixNano())
	}
	return tx.Save(u)
}

// Check if username or email is already in use
func (v *ItsAvailable) IsValid(errors *validate.Errors) {
	var qu User
	q := v.tx.Where("username = ?", v.Field)
	err := q.First(&qu)
	if err == nil {
		// user name already taken
		errors.Add(validators.GenerateKey(v.Name), fmt.Sprintf("The username %s is not available", v.Field))
		return
	}
	q = v.tx.Where("email = ?", v.Field2)
	err = q.First(&qu)
	if err == nil {
		// email is already taken
		errors.Add(validators.GenerateKey(v.Name), fmt.Sprintf("The email %s is not available", v.Field2))
		return
	}
}

// String is not required by pop and may be deleted
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Users is not required by pop and may be deleted
type Users []User

// String is not required by pop and may be deleted
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: u.Name, Name: "Name"},
		&validators.StringIsPresent{Field: u.Username, Name: "Username"},
		&validators.EmailIsPresent{Field: u.Email, Name: "Email"},
		&validators.StringIsPresent{Field: u.Password, Name: "Password"},
		&validators.StringsMatch{Field: u.Password, Field2: u.PasswordConfirm, Name: "PasswordConfirm", Message: "Passwords do not match"},
		&ItsAvailable{Field: u.Username, Field2: u.Email, tx: tx},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
