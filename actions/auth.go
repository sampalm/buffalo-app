package actions

import (
	"database/sql"
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/pkg/errors"
	"github.com/sampalm/buffalo/blogapp/models"
)

func init() {
	gothic.Store = App().SessionStore

	goth.UseProviders(
		github.New("67143c95c00aea5bdca3", "ef128240ed83559a2d2f1a1963d1b45d265aa7cb", fmt.Sprintf("%s%s", App().Host, "/auth/github/callback")),
	)
}

func AuthCallback(c buffalo.Context) error {
	guser, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return c.Error(401, err)
	}

	//return c.Render(200, r.JSON(guser))
	// Do somethingwith the user, maybe register them/sign them in
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}

	// Check for the user into the DB
	user := &models.User{}
	if err := tx.Where("provider = ? AND provider_id = ?", guser.Provider, guser.UserID).First(user); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			// Register isnt registered yet
			user = &models.User{
				Name:       guser.Name,
				Username:   guser.NickName,
				Email:      guser.Email,
				Admin:      false,
				Provider:   guser.Provider,
				ProviderID: guser.UserID,
			}
			if err = user.OAuthAndSave(tx); err != nil {
				return errors.WithStack(err)
			}
			c.Flash().Add("success", "User was created successfully")
			c.Redirect(302, "/login")
		}
		return errors.WithStack(err)
	}

	// Log in user
	c.Flash().Add("success", fmt.Sprintf("Hello %s, Welcome back!", user.Name))
	c.Session().Set("current_user_id", user.ID)
	return c.Redirect(302, "/")
}
