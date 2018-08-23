package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/pkg/errors"
	"github.com/sampalm/buffalo/blogapp/models"
)

// CommentsCreate POST default implementation.
func CommentsCreatePost(c buffalo.Context) error {
	// Get current user
	user := c.Value("current_user").(*models.User)
	// Bind Comments to the html form template
	comment := &models.Comment{}
	if err := c.Bind(comment); err != nil {
		return errors.WithStack(err)
	}
	// Get the DB connection from context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}
	// Get the Post from the parameter pid
	postID, err := uuid.FromString(c.Param("pid"))
	if err != nil {
		return errors.WithStack(err)
	}
	comment.PostID = postID
	comment.AuthorID = user.ID
	// Try to create the comment
	verrs, err := tx.ValidateAndCreate(comment)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Flash().Add("danger", "There was an error adding your comment.")
		return c.Redirect(302, "/posts/detail/%s", postID)
	}
	c.Flash().Add("success", "Comment added successfully.")
	return c.Redirect(302, "/posts/detail/%s", postID)
}

// CommentsEdit default implementation.
func CommentsEdit(c buffalo.Context) error {
	return c.Render(200, r.HTML("comments/edit.html"))
}

// CommentsDelete default implementation.
func CommentsDelete(c buffalo.Context) error {
	return c.Render(200, r.HTML("comments/delete.html"))
}
