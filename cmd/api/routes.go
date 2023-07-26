package main

import "net/http"

func (app *app) routes() http.Handler {
	app.Router.HandlerFunc(http.MethodGet, "/v1/buddy/user/:id", app.listBuddiesHandler)
	app.Router.HandlerFunc(http.MethodPost, "/v1/buddy", app.createBuddyHandler)

	app.Router.HandlerFunc(http.MethodPost, "/v1/diver", app.createDiverHandler)

	return app.Metrics(app.RecoverPanic(app.Router))
}
