package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/pkg/errors"
	"github.com/sampalm/buffalo/blogapp/models"
)

// TagsShow GET implementation.
func TagsShow(c buffalo.Context) error {
	// Get the DB connection from context.
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}

	// Get tag param from url
	tag := &models.Tag{}
	if err := tx.Where("code = ?", c.Param("tag")).First(tag); err != nil {
		return errors.WithStack(err)
	}

	// Set pagination
	posts := &models.Posts{}

	q := tx.PaginateFromParams(c.Params())
	err := q.Where("posts.id = tags_posts.post_id").LeftJoin("tags_posts", "tags_posts.tag_id = ?", tag.ID).All(posts)
	if err != nil {
		return errors.WithStack(err)
	}

	// Make posts available inside the html template
	c.Set("posts", posts)
	// Add pagination to the html
	c.Set("pagination", q.Paginator)

	return c.Render(200, r.HTML("posts/tags.html"))
}

// TagsCreate GET implementation
func TagsCreateGet(c buffalo.Context) error {
	c.Set("tag", &models.Tag{})
	return c.Render(200, r.HTML("posts/tags-create.html"))
}

// TagsCreate POST implementation
func TagsCreatePost(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}

	// Bind tag to the html form elements
	tag := &models.Tag{}
	if err := c.Bind(tag); err != nil {
		return errors.WithStack(err)
	}

	verrs, err := tag.Generate(tx)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("tag", tag)
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("posts/tags-create.html"))
	}

	c.Flash().Add("success", "A new tag was created successfully.")
	return c.Redirect(302, "/tags/list")
}

// TagsList Default implementation.
func TagsList(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("transaction not found"))
	}

	// Set pagination
	tags := &models.Tags{}
	q := tx.PaginateFromParams(c.Params())
	if err := q.All(tags); err != nil {
		return errors.WithStack(err)
	}

	// Bind results to the html template
	c.Set("tags", tags)
	c.Set("pagination", q.Paginator)

	return c.Render(200, r.HTML("posts/tags-list.html"))
}

// TagsDestryoy DELETE implementation
func TagsDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	tag := &models.Tag{}

	// Find the Tag using the code parameter
	if err := tx.Where("code = ?", c.Param("tag")).First(tag); err != nil {
		return c.Error(404, err)
	}

	// Delete Tag from DB
	if err := tx.Destroy(tag); err != nil {
		return errors.WithStack(err)
	}

	c.Flash().Add("success", "Tag as destroyed successfully")
	return c.Redirect(302, "/tags/list")
}
