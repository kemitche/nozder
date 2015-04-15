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
	host           *string
	port           *int
	templateDir    *string
	templates      *template.Template // Configured in initialize
	twitchClientID *string
	twitchCfg      *TwitchConfig // Configured in initialize
}

// TODO: Re-initialize on SIGHUP?
// TODO: panic on misconfig?
func (globals *Globals) initialize() {
	globals.templates = makeTemplates(*globals.templateDir)
	globals.twitchCfg = &TwitchConfig{version: 3, clientID: *globals.twitchClientID}
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
	globals.twitchClientID = flag.String("twitch-client-id", "", "Twitch API client ID")

	iniflags.Parse()

	fmt.Println(globals)

	return globals
}

type twitchPage struct {
	StreamID string
	Width    int
	Height   int
}

type pageList struct {
	TwitchPages [9]twitchPage
}

var requiredTemplates = []string{"twitch.html", "grid.html"}

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

func renderError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func renderTemplate(templates *template.Template, w http.ResponseWriter, templateName string, p interface{}) {
	err := templates.ExecuteTemplate(w, templateName, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func globalsMiddleware(globals *Globals) func(*Context, web.ResponseWriter, *web.Request, web.NextMiddlewareFunc) {
	return func(c *Context, rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
		c.globals = globals
		next(rw, req)
	}
}

func (c *Context) showTwitchStream(rw web.ResponseWriter, req *web.Request) {
	page := &twitchPage{StreamID: req.PathParams["id"], Width: 800, Height: 600}
	renderTemplate(c.globals.templates, rw, "twitch.html", page)
}

func (c *Context) showTwitchGrid(rw web.ResponseWriter, req *web.Request) {
	streamID := req.PathParams["id"]
	streams := pageList{}
	for i := 0; i < len(streams.TwitchPages); i++ {
		streams.TwitchPages[i].Height = 300
		streams.TwitchPages[i].Width = 300
		streams.TwitchPages[i].StreamID = streamID
	}
	renderTemplate(c.globals.templates, rw, "grid.html", streams)
}

func (c *Context) tempSearch(rw web.ResponseWriter, req *web.Request) {
	query := req.PathParams["query"]
	var cfg TwitchConfig = c.globals.twitchCfg
	cfg.searchForGameStreams(query, false)
}

func setUpRoutes(router *web.Router) {
	router.Get("/twitch/:id", (*Context).showTwitchStream)
	router.Get("/grid/:id", (*Context).showTwitchGrid)
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
