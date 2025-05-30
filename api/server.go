package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/flemeur/taleplade"
	"github.com/flemeur/taleplade/errors"
	"github.com/flemeur/taleplade/memcache"
)

func NewServer(config *taleplade.Config, cache *memcache.Cache) http.Handler {
	r := chi.NewRouter()

	r.MethodNotAllowed(HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		return Respond(w, http.StatusMethodNotAllowed, errorResponse{
			Error:   http.StatusText(http.StatusMethodNotAllowed),
			Message: "The method is not allowed",
		})
	}))
	r.NotFound(HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		return errors.E(errors.NotExists, "The URL could not be found")
	}))

	s := &server{
		config: config,
		cache:  cache,
		router: r,
	}

	s.attachRoutes()

	return s
}

type server struct {
	config *taleplade.Config
	cache  *memcache.Cache
	router chi.Router
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hErr := h(w, r)
	if hErr == nil {
		return
	}

	if err := RespondErr(w, hErr); err != nil {
		log.Printf("api.Handler.ServeHTTP: RespondErr: %v", err)
	}

	if errors.IsKind(errors.Internal, hErr) {
		log.Printf("api.Handler.ServeHTTP: %v", hErr)
	}
}

func HandlerFunc(h Handler) http.HandlerFunc { return h.ServeHTTP }

func Respond(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return errors.E(fmt.Errorf("json.NewEncoder().Encode: %w", err))
	}

	return nil
}

func RespondErr(w http.ResponseWriter, err error) error {
	status := errorKindToStatusCode(err)

	return Respond(w, status, errorResponse{
		Error:   http.StatusText(status),
		Message: errors.ErrorMessage(err),
	})
}

func errorKindToStatusCode(err error) int {
	switch errors.ErrorKind(err) {
	case errors.Authentication:
		return http.StatusUnauthorized
	case errors.Permission:
		return http.StatusForbidden
	case errors.Invalid:
		return http.StatusBadRequest
	case errors.Exists:
		return http.StatusConflict
	case errors.NotExists:
		return http.StatusNotFound
	case errors.Validation:
		return http.StatusUnprocessableEntity
	case errors.Transient:
		return http.StatusServiceUnavailable
	case errors.External:
		return http.StatusBadGateway
	case errors.Timeout:
		return http.StatusGatewayTimeout

	case errors.Internal:
	case errors.Undefined:
	}

	return http.StatusInternalServerError
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
