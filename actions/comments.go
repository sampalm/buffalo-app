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

// CommentsEdit GET default implementation.
func CommentsEditGet(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}

	// Get comments from current user
	comment := &models.Comment{}
	user := c.Value("current_user").(*models.User)
	if err := tx.Find(comment, c.Param("cid")); err != nil {
		return c.Error(404, err)
	}

	// Make sure the Author is the logged in user
	if user.ID != comment.AuthorID {
		c.Flash().Add("danger", "You are not authorized to view that page")
		return c.Redirect(302, "/posts/detail/%s", comment.PostID)
	}

	c.Set("comment", comment)
	return c.Render(200, r.HTML("comments/edit.html"))
}

// CommentsEdit POST default implementation.
func CommentsEditPost(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}

	// Get comments from the parameter cid
	comment := &models.Comment{}
	if err := tx.Find(comment, c.Param("cid")); err != nil {
		return c.Error(404, err)
	}

	// Bind the comments to the html page
	if err := c.Bind(comment); err != nil {
		return errors.WithStack(err)
	}

	// Make sure the Author is the logged in user
	user := c.Value("current_user").(*models.User)
	if user.ID != comment.AuthorID {
		c.Flash().Add("danger", "You are not authorized to view that page.")
		return c.Redirect(302, "/posts/detail/%s", comment.PostID)
	}

	// Update the comment in DB
	verrs, err := tx.ValidateAndUpdate(comment)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("comment", comment)
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("comments/edit.html"))
	}

	c.Flash().Add("success", "Comment was updated successfully")
	return c.Redirect(302, "/posts/detail/%s", comment.PostID)
}

// CommentsDelete default implementation.
func CommentsDelete(c buffalo.Context) error {
	// Get the DB connection from context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}

	// Get comment from the parameter cid
	comment := &models.Comment{}
	if err := tx.Find(comment, c.Param("cid")); err != nil {
		return c.Error(404, err)
	}

	// Check if the user is the Author or Admin
	user := c.Value("current_user").(*models.User)
	if user.ID != comment.PostID && user.Admin == false {
		c.Flash().Add("danger", "You are not authorized to view that page")
		return c.Redirect(302, "/posts/detail/%s", comment.PostID)
	}

	// Delete comment from DB
	err := tx.Destroy(comment)
	if err != nil {
		return errors.WithStack(err)
	}

	c.Flash().Add("success", "Comment deleted successfully")
	return c.Redirect(302, "/posts/detail/%s", comment.PostID)
}
