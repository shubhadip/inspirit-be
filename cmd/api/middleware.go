package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pascaldekloe/jwt"
)

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		next.ServeHTTP(w, r)
	})
}

func (app *application) checkToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			// could set an anony. user
		}

		headerParts := strings.Split(authHeader, " ")

		if len(headerParts) != 2 {
			app.errorJSON(w, errors.New("Invalid auth Header"), http.StatusForbidden)
			return
		}

		if headerParts[0] != "Bearer" {
			app.errorJSON(w, errors.New("unAuthorised no bearer"), http.StatusForbidden)
			return
		}

		token := headerParts[1]
		claims, err := jwt.HMACCheck([]byte(token), []byte(app.config.jwt.secret))

		if err != nil {
			app.errorJSON(w, errors.New("unAuthorised hmac"), http.StatusUnauthorized)
			return
		}

		if !claims.Valid(time.Now()) {
			app.errorJSON(w, errors.New("token expired"), http.StatusUnauthorized)
			return
		}

		if !claims.AcceptAudience("mydomain.com") {
			app.errorJSON(w, errors.New("invalid audience"), http.StatusUnauthorized)
			return
		}

		if claims.Issuer != "mydomain.com" {
			app.errorJSON(w, errors.New("invalid issuer"), http.StatusUnauthorized)
			return
		}

		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			app.errorJSON(w, errors.New("unAuthorised"), http.StatusUnauthorized)
			return
		}

		log.Println("Valid User", userID)
		next.ServeHTTP(w, r)
	})
}
