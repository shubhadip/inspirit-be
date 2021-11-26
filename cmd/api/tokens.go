package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pascaldekloe/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `json: "username"`
	Password string `json: "password"`
}

func (app *application) signIn(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		app.errorJSON(w, errors.New("unAuthorized"))
		return
	}

	user, err := app.models.DB.GetUser(creds.Username)

	if err != nil {
		app.errorJSON(w, errors.New("User not registered"))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))
	if err != nil {
		app.errorJSON(w, errors.New("unAuthorized"))
		return
	}

	var claims jwt.Claims

	claims.Subject = fmt.Sprint(user.ID)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(3 * time.Hour))
	claims.Issuer = "mydomain.com"
	claims.Audiences = []string{"mydomain.com"}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secret))

	if err != nil {
		app.errorJSON(w, errors.New("Error Sign In"))
		return
	}

	app.writeJSON(w, http.StatusOK, string(jwtBytes), "response")

}

func (app *application) validateToken(w http.ResponseWriter, r *http.Request) {
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
	app.writeJSON(w, http.StatusOK, string("Valid Token"), "response")
}
