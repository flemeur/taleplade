package api

import (
	"net/http"
	"time"
)

func (s *server) attachRoutes() {
	s.router.Method(http.MethodGet, "/", Handler(func(w http.ResponseWriter, r *http.Request) error {
		return Respond(w, http.StatusOK, struct {
			Message string `json:"message"`
		}{Message: "Hello"})
	}))

	s.router.Method(http.MethodGet, "/phrases", Handler(s.handleGetPhrases))
	s.router.Method(http.MethodGet, "/names", Handler(s.handleGetNames))

	s.router.Method(http.MethodGet, "/tts", CacheFor(14*24*time.Hour)(Handler(s.handleTTS)))
}
