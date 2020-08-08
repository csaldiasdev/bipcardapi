package main

import (
	"bipcardapi/way"
	"log"
	"net/http"
)

const baseURL = "http://pocae.tstgo.cl/PortalCAE-WAR-MODULE"

func main() {
	router := way.NewRouter()
	router.HandleFunc("GET", "/api/bipcard/:cardId/balance", handleBipCardBalance)
	router.HandleFunc("GET", "/api/bipcard/:cardId/movements", handleBipCardMovements)
	log.Fatalln(http.ListenAndServe(":8080", router))
}

func handleBipCardBalance(w http.ResponseWriter, r *http.Request) {
	cardID := way.Param(r.Context(), "cardId")
	log.Println(cardID)
}

func handleBipCardMovements(w http.ResponseWriter, r *http.Request) {
	cardID := way.Param(r.Context(), "cardId")
	log.Println(cardID)
}
