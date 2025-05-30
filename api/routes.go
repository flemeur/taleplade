package api

import (
	"net/http"
)

func (s *server) attachRoutes() {
	s.router.Method(http.MethodGet, "/", Handler(func(w http.ResponseWriter, r *http.Request) error {
		return Respond(w, http.StatusOK, struct {
			Message string `json:"message"`
		}{Message: "Hello there!"})
	}))

	// s.router.Method(http.MethodGet, "/tts/{language}", CacheFor(30*24*time.Hour)(Handler(s.handleTextToSpeech)))

	s.router.Method(http.MethodGet, "/phrases", Handler(s.handleGetPhrases))
	s.router.Method(http.MethodGet, "/names", Handler(s.handleGetNames))
}
