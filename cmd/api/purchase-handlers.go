package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/pascaldekloe/jwt"
)

type buyPayload struct {
	Amount float64 `json:"amount"`
}

type sellPayload struct {
	Amount       float64 `json:"amount"`
	BitcoinValue float64 `json: "bitcoin_value"`
}

type Inner struct {
	Id                string `json: "id"`
	Rank              string `json: "rank"`
	Symbol            string `json: "symbol"`
	Name              string `json: "name"`
	Supply            string `json: "supply"`
	MaxSupply         string `json: "maxSupply"`
	MarketCapUsd      string `json: "marketCapUsd"`
	VolumeUsd24Hr     string `json: "volumeUsd24Hr"`
	PriceUsd          string `json: "priceUsd"`
	ChangePercent24Hr string `json: "changePercent24Hr"`
	Vwap24Hr          string `json: "vwap24Hr"`
	Explorer          string `json: "explorer"`
}

type Outmost struct {
	Key Inner `json:"data"`
}

func getUserIdFromToken(app *application, w http.ResponseWriter, r *http.Request) int64 {

	w.Header().Add("Vary", "Authorization")
	authHeader := r.Header.Get("Authorization")
	headerParts := strings.Split(authHeader, " ")
	token := headerParts[1]
	claims, _ := jwt.HMACCheck([]byte(token), []byte(app.config.jwt.secret))
	userID, _ := strconv.ParseInt(claims.Subject, 10, 64)
	return userID
}

func getBitcoinAmount() float64 {
	response, err := http.Get("https://api.coincap.io/v2/assets/bitcoin")
	var f float64

	if err != nil {
		log.Println("Http request failed", err)
		f = 57793.6073427724428628
	}

	byteSlice, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	var cont Outmost
	json.Unmarshal([]byte(byteSlice), &cont)

	f, err = strconv.ParseFloat(cont.Key.PriceUsd, 16)

	if err != nil {
		log.Println("Parsing Failed", err)
		f = 57793.6073427724428628
	}
	log.Println(f)
	return f
}

func (app *application) buy(w http.ResponseWriter, r *http.Request) {
	var payload buyPayload

	errDecode := json.NewDecoder(r.Body).Decode(&payload)
	if errDecode != nil {
		log.Println(errDecode)
		app.errorJSON(w, errDecode)
		return
	}

	userID := getUserIdFromToken(app, w, r)
	currentBitCoinValue := getBitcoinAmount()

	user, err := app.models.DB.PurchaseBitCoin(userID, payload.Amount, currentBitCoinValue)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, user, "response")
	if err != nil {
		app.errorJSON(w, err)
		return
	}

}

func (app *application) sell(w http.ResponseWriter, r *http.Request) {
	var payload sellPayload

	errDecode := json.NewDecoder(r.Body).Decode(&payload)
	if errDecode != nil {
		log.Println(errDecode)
		app.errorJSON(w, errDecode)
		return
	}
	userID := getUserIdFromToken(app, w, r)
	currentBitCoinValue := getBitcoinAmount()

	user, err := app.models.DB.SellBitCoin(userID, payload.Amount, currentBitCoinValue)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, user, "response")
	if err != nil {
		app.errorJSON(w, err)
		return
	}

}
