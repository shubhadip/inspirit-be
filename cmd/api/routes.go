package main

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) wrap(next http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		//pass httprouter.Params to request context
		ctx := context.WithValue(r.Context(), httprouter.ParamsKey, ps)
		//call next middleware with new context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (app *application) routes() http.Handler {
	router := httprouter.New()

	secure := alice.New(app.checkToken)

	router.HandlerFunc(http.MethodPost, "/api/login", app.signIn)
	router.HandlerFunc(http.MethodPost, "/api/validate", app.validateToken)
	router.HandlerFunc(http.MethodPost, "/api/signup", app.createUser)
	router.GET("/api/getuser/:id", app.wrap(secure.ThenFunc(app.getUser)))
	router.POST("/api/buy", app.wrap(secure.ThenFunc(app.buy)))
	router.POST("/api/sell", app.wrap(secure.ThenFunc(app.sell)))
	// router.HandlerFunc(http.MethodGet, "/api/bitcoin", app.getBitcoinAmount)

	return app.enableCORS(router)
}
