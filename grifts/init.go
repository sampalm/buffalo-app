package grifts

import (
	"github.com/gobuffalo/buffalo"
	"github.com/sampalm/buffalo/blogapp/actions"
)

func init() {
	buffalo.Grifts(actions.App())
}
