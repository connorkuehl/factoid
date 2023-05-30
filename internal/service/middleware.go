package service

import (
	"errors"
	"net/http"
)

func (s *Service) privileged(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.auth != "" && r.Header.Get("Authorization") != s.auth {
			s.RespondErrorJSON(w, http.StatusForbidden, errors.New("forbidden"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
