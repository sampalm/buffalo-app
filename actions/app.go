package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/ssl"
	"github.com/gobuffalo/envy"
	"github.com/unrolled/secure"

	"github.com/gobuffalo/buffalo/middleware/csrf"
	"github.com/gobuffalo/buffalo/middleware/i18n"
	"github.com/gobuffalo/packr"
	"github.com/sampalm/buffalo/blogapp/models"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:         ENV,
			SessionName: "_blogapp_session",
		})
		// Automatically redirect to SSL
		app.Use(forceSSL())

		if ENV == "development" {
			app.Use(middleware.ParameterLogger)
		}

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(csrf.New)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.PopTransaction)
		// Remove to disable this.
		app.Use(middleware.PopTransaction(models.DB))

		// Save current user into context
		app.Use(SetCurrentUser)

		// Setup and use translations:
		app.Use(translations())

		app.GET("/", HomeHandler)

		// Users routing
		users := app.Group("/users")
		users.GET("/", AdminRequired(List))
		users.POST("/", Create)
		users.GET("/new", New)
		users.GET("/{user_id}", Show)
		users.PUT("/{user_id}", Update)
		users.DELETE("/{user_id}", AdminRequired(Destroy))
		users.GET("/{user_id}/edit", Edit)
		app.GET("/login", UsersLogin)
		app.POST("/login", UsersLoginPost)
		app.GET("/logout", UsersLogout)

		// Posts routing
		posts := app.Group("/posts")
		posts.GET("/", PostsIndex)
		posts.GET("/create", AdminRequired(PostsCreateGet))
		posts.POST("/create", AdminRequired(PostsCreatePost))
		posts.GET("/edit/{pid}", AdminRequired(PostsEditGet))
		posts.POST("/edit/{pid}", AdminRequired(PostsEditPost))
		posts.GET("/delete/{pid}", PostsDelete)
		posts.GET("/detail/{pid}", PostsDetail)

		// Comments routing
		comments := app.Group("/comments")
		comments.Use(LoginRequired)
		comments.POST("/create/{pid}", CommentsCreatePost)
		comments.GET("/edit", CommentsEdit)
		comments.GET("/delete", CommentsDelete)

		app.ServeFiles("/", assetsBox) // serve files from the public directory
	}

	return app
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(packr.NewBox("../locales"), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return ssl.ForceSSL(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}
