package main

import (
	"backend/models"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

type jsonResp struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type userPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	var payload userPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Println(err)
		app.errorJSON(w, err)
		return
	}

	var user models.User

	hash, _ := HashPassword(payload.Password) // ignore error for the sake of simplicity
	user.Username = payload.Username
	user.Password = hash
	user.WalletAmount = 500000
	user.BitcoinValue = 0
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	err = app.models.DB.InsertUser(user)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	ok := jsonResp{
		OK:      true,
		Message: "User created SuccessFully",
	}

	err = app.writeJSON(w, http.StatusOK, ok, "response")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}

func (app *application) getUser(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	app.logger.Print(params)
	id, err := strconv.Atoi(params.ByName("id"))
	app.logger.Print(id)
	if err != nil {
		app.logger.Print(errors.New("invalid id parameter"))
		app.errorJSON(w, err)
		return
	}

	user, err := app.models.DB.GetUserById(id)

	err = app.writeJSON(w, http.StatusOK, user, "user")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}
