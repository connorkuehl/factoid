package service

import (
	"encoding/json"
	"net/http"
)

func (s *Service) RespondJSON(w http.ResponseWriter, code int, envelope map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(envelope); err != nil {
		s.logger.Error().Err(err).Msg("")
	}
}

func (s *Service) RespondErrorJSON(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]any{"error": err.Error()}); err != nil {
		s.logger.Error().Err(err).Msg("")
	}
}
