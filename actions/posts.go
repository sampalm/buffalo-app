package actions

import (
	"strconv"

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
	tx := c.Value("tx").(*pop.Connection)

	tags := &models.Tags{}
	if err := tx.All(tags); err != nil {
		errors.WithStack(err)
	}

	c.Set("tags", tags)
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

	// Get FileImage from html form
	f, err := c.File("FileImage")
	if err != nil {
		return errors.WithStack(err)
	}
	post.FileImage = f

	// Validate the data from html form
	post.AuthorID = user.ID
	veers, err := post.UploadAndCreate(tx)
	if err != nil {
		return errors.WithStack(err)
	}

	if veers.HasAny() {
		c.Set("post", post)
		c.Set("errors", veers.Errors)
		return c.Render(422, r.HTML("posts/create"))
	}

	// Validate the posts tag
	if post.Tag != "" {
		tag := &models.Tag{}
		code, err := strconv.Atoi(post.Tag)
		if err != nil {
			return errors.WithStack(err)
		}
		if err := tx.Where("code = ?", code).First(tag); err != nil {
			return errors.WithStack(err)
		}
		tagpost := &models.TagPost{}
		tagpost.PostID = post.ID
		tagpost.TagID = tag.ID
		if err := tx.Save(tagpost); err != nil {
			return errors.WithStack(err)
		}
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

	// Get Post from DB to html template
	post := &models.Post{}
	if err := tx.Find(post, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	// Get Tag from DB to html template
	tag := &models.Tag{}
	err := tx.Q().Where("tags.id = tags_posts.tag_id").LeftJoin("tags_posts", "tags_posts.post_id = ?", post.ID).First(tag)
	if err != nil {
		return c.Error(404, err)
	}
	// Get All Tags from DB to html template
	tags := &models.Tags{}
	if err := tx.All(tags); err != nil {
		return c.Error(404, err)
	}

	c.Set("post", post)
	c.Set("post_tag", tag)
	c.Set("tags", tags)
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
	if err := tx.Find(post, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}

	// Bind post to the html form element
	if err := c.Bind(post); err != nil {
		return errors.WithStack(err)
	}

	// Get file from file_input form
	f, err := c.File("FileImage")
	if err != nil {
		return errors.WithStack(err)
	}
	post.FileImage = f

	// Try to update post data in the DB
	verrs, err := post.UploadAndUpdated(tx)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("post", post)
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("posts/edit.html"))
	}

	// Try to update post tags
	pTag := &models.TagPost{}
	err = tx.Q().LeftJoin("tags", "tags_posts.post_id = ?", post.ID).Where("tags.id = tags_posts.tag_id").First(pTag)
	if err != nil {
		return errors.WithStack(err)
	}
	newTag := &models.Tag{}
	if err = tx.Where("code = ?", post.Tag).First(newTag); err != nil {
		return errors.WithStack(err)
	}
	pTag.PostID = post.ID
	pTag.TagID = newTag.ID
	if err = tx.Update(pTag); err != nil {
		return errors.WithStack(err)
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "Post was updated successfully.")
	return c.Redirect(302, "/posts/detail/%s", post.ID)
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

	// Try to exclude image from disk
	if err := post.DeleteFile(tx); err != nil {
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

	// Find the posts tags
	tags := &models.Tags{}
	err := tx.Q().Where("tags.id = tags_posts.tag_id").LeftJoin("tags_posts", "tags_posts.post_id = ?", post.ID).All(tags)
	if err != nil {
		return c.Error(404, err)
	}

	// Bind Post content and Author to html template
	c.Set("post", post)
	c.Set("author", author)
	c.Set("tags", tags)

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
