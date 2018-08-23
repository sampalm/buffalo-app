package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/pkg/errors"
	"github.com/sampalm/buffalo/blogapp/models"
)

// PostsIndex default implementation.
func PostsIndex(c buffalo.Context) error {
	// Get the DB connection from contect
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}
	posts := &models.Posts{}

	// Set paginate results. Params "page" and "per_page" control pagination.
	q := tx.PaginateFromParams(c.Params())
	// Query all Posts from the DB
	if err := q.All(posts); err != nil {
		return errors.WithStack(err)
	}

	// Make posts available inside the html template
	c.Set("posts", posts)
	// Add the paginator to the context so it can be used in the html
	c.Set("pagination", q.Paginator)

	return c.Render(200, r.HTML("posts/index.html"))
}

// PostsCreate GET implementation.
func PostsCreateGet(c buffalo.Context) error {
	c.Set("post", &models.Post{})
	return c.Render(200, r.HTML("posts/create.html"))
}

// PostsCreate POST implementation.
func PostsCreatePost(c buffalo.Context) error {
	// Allocate an empty post and get user
	post := &models.Post{}
	user := c.Value("current_user").(*models.User)

	// Bind post to the html form elements
	if err := c.Bind(post); err != nil {
		return errors.WithStack(err)
	}

	// Get the DB connection from context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}

	// Validate the data from html form
	post.AuthorID = user.ID
	veers, err := tx.ValidateAndCreate(post)
	if err != nil {
		return errors.WithStack(err)
	}

	if veers.HasAny() {
		c.Set("post", post)
		c.Set("errors", veers.Errors)
		return c.Render(422, r.HTML("post/create"))
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "New post added successfully")
	return c.Redirect(302, "/posts")
}

// PostsEdit GET implementation.
func PostsEditGet(c buffalo.Context) error {
	// Get the DB connection from context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}

	// Bind Post to html template
	post := &models.Post{}
	if err := tx.Find(post, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}

	c.Set("post", post)
	return c.Render(200, r.HTML("posts/edit.html"))
}

// PostsEdit POST implementation.
func PostsEditPost(c buffalo.Context) error {
	// Get th DB connection from context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}

	// To find the Post the parameter pid is used
	post := &models.Post{}
	if err := tx.Find(post, c.Param("pdi")); err != nil {
		return c.Error(404, err)
	}

	// Bind post to the html form element
	if err := c.Bind(post); err != nil {
		return errors.WithStack(err)
	}

	// Try to update post data in the DB
	verrs, err := tx.ValidateAndUpdate(post)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("post", post)
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("posts/edit.html"))
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "Post was updated successfully.")
	return c.Redirect(302, "posts/detail/%s", post.ID)
}

// PostsDelete default implementation.
func PostsDelete(c buffalo.Context) error {
	// Get the DB connection form context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}

	// Try to find post in the trasaction using pid parameter
	post := &models.Post{}
	if err := tx.Find(post, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}

	// Try to exclude post from DB
	if err := tx.Destroy(post); err != nil {
		return errors.WithStack(err)
	}

	c.Flash().Add("success", "Post was successfully deleted.")
	return c.Redirect(302, "/posts")
}

// PostsDetail default implementation.
func PostsDetail(c buffalo.Context) error {
	// Get the DB connnection from context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}

	// To find the Post the parameter pid is used
	post := &models.Post{}
	if err := tx.Find(post, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}

	// To find the Post Author the parameter AuthorID is used
	author := &models.User{}
	if err := tx.Find(author, post.AuthorID); err != nil {
		return c.Error(404, err)
	}

	// Bind Post content and Author to html template
	c.Set("post", post)
	c.Set("author", author)

	// Get the comments for this posts
	comment := &models.Comment{}
	c.Set("comment", comment)
	comments := models.Comments{}
	if err := tx.BelongsTo(post).All(&comments); err != nil {
		return errors.WithStack(err)
	}

	// To find the Comments Author
	for i := 0; i < len(comments); i++ {
		u := models.User{}
		if err := tx.Find(&u, comments[i].AuthorID); err != nil {
			return c.Error(404, err)
		}
		comments[i].Author = u
	}
	c.Set("comments", comments)
	return c.Render(200, r.HTML("posts/detail.html"))
}
