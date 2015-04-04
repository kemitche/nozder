package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gocraft/web"
)

// Context is the Nozder web app's request context
type Context struct {
	globals *Globals
}

// Globals exist through every request; this should be configuration mostly
type Globals struct {
	templates *template.Template
}

func makeTemplates(templateDir string) *template.Template {
	return nil
}

func globalsMiddleware(templateDir string) func(*Context, web.ResponseWriter, *web.Request, web.NextMiddlewareFunc) {
	globals := new(Globals)
	globals.templates = makeTemplates(templateDir)
	return func(c *Context, rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
		c.globals = globals
		next(rw, req)
	}
}

func (c *Context) showTwitchStream(rw web.ResponseWriter, req *web.Request) {
	fmt.Fprint(rw, "Hello world!")
	fmt.Fprint(rw, "<br>You're about to watch:", req.PathParams["id"])
}

func setUpRoutes(router *web.Router) {
	router.Get("/twitch/:id", (*Context).showTwitchStream)
}

func main() {
	router := web.New(Context{})
	router.Middleware(web.LoggerMiddleware)
	router.Middleware(globalsMiddleware(""))

	setUpRoutes(router)
	http.ListenAndServe("localhost:3000", router)
}
