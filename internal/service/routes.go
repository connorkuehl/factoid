package service

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (s *Service) Routes() *httprouter.Router {
	mux := httprouter.New()

	mux.HandlerFunc(http.MethodGet, "/v1/facts", s.FactsHandler)
	mux.HandlerFunc(http.MethodGet, "/v1/fact/:id", s.FactHandler)
	mux.HandlerFunc(http.MethodPost, "/v1/facts", s.FactsHandler)
	mux.HandlerFunc(http.MethodDelete, "/v1/fact/:id", s.FactHandler)

	return mux
}
