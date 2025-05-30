package api

import "net/http"

func (s *server) handleGetPhrases(w http.ResponseWriter, r *http.Request) error {
	return Respond(w, http.StatusOK, s.config.Phrases)
}

func (s *server) handleGetNames(w http.ResponseWriter, r *http.Request) error {
	return Respond(w, http.StatusOK, s.config.Names)
}
