package service

import (
	"encoding/json"
	"net/http"

	log "golang.org/x/exp/slog"
)

func (s *Service) RespondJSON(w http.ResponseWriter, code int, envelope map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(envelope); err != nil {
		log.With("err", err).Error("")
	}
}

func (s *Service) RespondErrorJSON(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]any{"error": err.Error()}); err != nil {
		log.With("err", err).Error("")
	}
}
