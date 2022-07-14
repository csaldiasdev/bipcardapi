package server

import (
	"bipcardapi/internal/bipcard"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

const cardNumberKeyParam = "CARD_NUMBER"

type server struct {
	bipCardClient bipcard.BipCardClient
}

func newServer() server {
	return server{
		bipCardClient: bipcard.NewBipCardClient(),
	}
}

func NewHTTPServer(addr string) *http.Server {
	srv := newServer()
	r := mux.NewRouter()

	r.HandleFunc("/api/v1/bipcard/{"+cardNumberKeyParam+"}/info", srv.handleCardInfo).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/bipcard/{"+cardNumberKeyParam+"}/movements", srv.handleCardMovements).Methods(http.MethodGet)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

func (s server) handleCardInfo(w http.ResponseWriter, r *http.Request) {
	cNumber := getPathParam(r, cardNumberKeyParam)

	data, err := s.bipCardClient.GetBipCardInfo(cNumber)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(data)
}

func (s server) handleCardMovements(w http.ResponseWriter, r *http.Request) {
	cNumber := getPathParam(r, cardNumberKeyParam)

	data, err := s.bipCardClient.GetBipCardMovements(cNumber)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(data)
}

func getPathParam(r *http.Request, key string) string {
	vars := mux.Vars(r)
	return vars[key]
}
