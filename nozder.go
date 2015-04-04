package main

import (
	"fmt"
	"net/http"

	"github.com/gocraft/web"
)

// Context is the Nozder web app's request context
type Context struct {
}

func (c *Context) showTwitchStream(rw web.ResponseWriter, req *web.Request) {
	fmt.Fprint(rw, "Hello world!")
	fmt.Fprint(rw, "You're about to watch:", req.PathParams["id"])
}

func setUpRoutes(router *web.Router) {
	router.Get("/twitch/:id", (*Context).showTwitchStream)
}

func main() {
	router := web.New(Context{})
	setUpRoutes(router)
	http.ListenAndServe("localhost:3000", router)
}
