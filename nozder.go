package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gocraft/web"
	"github.com/vharitonsky/iniflags"
)

// Context is the Nozder web app's request context
type Context struct {
	globals *Globals
}

// Globals exist through every request; this should be configuration mostly
type Globals struct {
	host        *string
	port        *int
	templateDir *string
	templates   *template.Template
}

func (globals *Globals) initialize() {
	globals.templates = makeTemplates(*globals.templateDir)
}

func (globals *Globals) serveOn() string {
	return fmt.Sprintf("%s:%d", *globals.host, *globals.port)
}

func (globals *Globals) String() string {
	return fmt.Sprintf("Serving on: %s", globals.serveOn())
}

func makeGlobals() *Globals {
	globals := new(Globals)
	globals.host = flag.String("host", "localhost", "Server host name")
	globals.port = flag.Int("port", 9000, "Server listen port")
	globals.templateDir = flag.String("templates", "", "Location of HTML templates")

	iniflags.Parse()

	fmt.Println(globals)

	return globals
}

type twitchPage struct {
	streamID string
}

var requiredTemplates = []string{"twitch.html"}

func makeTemplates(templateDir string) *template.Template {
	if templateDir == "" {
		// Assume this was run from the root directory of nozder source
		cwd, err := os.Getwd()
		if err != nil {
			// We're not going to be able to load the templates if we don't have
			// a template directory.
			panic(err)
		}
		templateDir = filepath.Join(cwd, "templates")
	}
	// TODO: Better logging!
	fmt.Println("loading templates from " + templateDir)

	templatePaths := make([]string, len(requiredTemplates), len(requiredTemplates))
	for index, templateName := range requiredTemplates {
		templatePaths[index] = filepath.Join(templateDir, templateName)
	}

	return template.Must(template.ParseFiles(templatePaths...))
}

func globalsMiddleware(globals *Globals) func(*Context, web.ResponseWriter, *web.Request, web.NextMiddlewareFunc) {
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
	globals := makeGlobals()
	globals.initialize()

	router := web.New(Context{})
	router.Middleware(web.LoggerMiddleware)
	router.Middleware(globalsMiddleware(globals))

	setUpRoutes(router)
	http.ListenAndServe(globals.serveOn(), router)
}
