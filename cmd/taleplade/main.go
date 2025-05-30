package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/flemeur/taleplade"
	"github.com/flemeur/taleplade/api"
	"github.com/flemeur/taleplade/memcache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lpar/gzipped/v2"
)

func run() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("os.Executable: %w", err)
	}
	execDir := filepath.Dir(executable)

	var config taleplade.Config
	configFile, err := os.Open(filepath.Join(execDir, "config.json"))
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return fmt.Errorf("json.NewDecoder: %w", err)
	}

	cache := memcache.New("memcached:11211")
	if err := cache.Ping(); err != nil {
		return fmt.Errorf("memcache.Ping: %w", err)
	}

	apiServer := api.NewServer(&config, cache)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Heartbeat("/health"))
	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	router.Use(middleware.Timeout(time.Minute))
	router.Use(middleware.Compress(5,
		"application/javascript",
		"application/json",
		"application/x-javascript",
		"image/svg+xml",
		"text/css",
		"text/html",
		"text/plain",
		"text/xml",
	))

	router.Mount("/api", apiServer)

	wwwDir := filepath.Join(execDir, "public")
	router.Mount("/", frontendHandler(wwwDir))

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      90 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Print(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-ch

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Print(err)
	}

	return nil
}

func frontendHandler(dir string) http.Handler {
	handler := withIndexHTML(gzipped.FileServer(gzipped.Dir(dir)))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path

		if strings.Contains(urlPath, ".") || urlPath == "/" {
			handler.ServeHTTP(w, r)

			return
		}

		// Fall back to index.html
		http.ServeFile(w, r, path.Join(dir, "index.html"))
	})
}

func withIndexHTML(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") || len(r.URL.Path) == 0 {
			newPath := path.Join(r.URL.Path, "index.html")
			r.URL.Path = newPath
		}

		h.ServeHTTP(w, r)
	})
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
