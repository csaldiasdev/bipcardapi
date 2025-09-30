package server

import (
	"bipcardapi/internal/bipcard"
	"encoding/json"
	"net/http"
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

	stdMux := http.NewServeMux()

	stdMux.HandleFunc("GET /api/v1/bipcard/{"+cardNumberKeyParam+"}/info", srv.handleCardInfo)
	stdMux.HandleFunc("GET /api/v1/bipcard/{"+cardNumberKeyParam+"}/movements", srv.handleCardMovements)

	return &http.Server{
		Addr:    addr,
		Handler: stdMux,
	}
}

func (s server) handleCardInfo(w http.ResponseWriter, r *http.Request) {
	cNumber := r.PathValue(cardNumberKeyParam)

	data, err := s.bipCardClient.GetBipCardInfo(cNumber)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(data)
}

func (s server) handleCardMovements(w http.ResponseWriter, r *http.Request) {
	cNumber := r.PathValue(cardNumberKeyParam)

	data, err := s.bipCardClient.GetBipCardMovements(cNumber)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(data)
}
