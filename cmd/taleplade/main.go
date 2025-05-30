package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/flemeur/taleplade"
	"github.com/flemeur/taleplade/api"
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

	apiServer := api.NewServer(&config)

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

	router.Mount("/tts", ttsProxy())

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

func ttsProxy() http.Handler {
	// "Free" TTS solution with good support for danish
	// https://translate.google.com/translate_tts?ie=UTF-8&client=tw-ob&q=test+af+tekst+til+tale&tl=da

	target, err := url.Parse("https://translate.google.com/translate_tts?ie=UTF-8&client=tw-ob&tl=da")
	if err != nil {
		panic(err)
	}

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			targetQuery := target.RawQuery
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = target.EscapedPath()
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
			req.Header.Del("Cookie")
			req.Header.Del("Referer")
			req.Host = "" // Set r.Host empty to use r.URL.Host instead
		},
		ModifyResponse: func(resp *http.Response) error {
			// Content-Type: audio/mpeg
			// Transfer-Encoding: chunked

			resp.Header.Del("Accept-Ch")
			resp.Header.Del("Alt-Svc")
			resp.Header.Del("Content-Security-Policy")
			resp.Header.Del("Content-Security-Policy")
			resp.Header.Del("Cross-Origin-Opener-Policy")
			resp.Header.Del("Cross-Origin-Resource-Policy")
			resp.Header.Del("P3p")
			resp.Header.Del("Permissions-Policy")
			resp.Header.Del("Reporting-Endpoints")
			resp.Header.Del("Server")
			resp.Header.Del("X-Content-Type-Options")
			resp.Header.Del("X-Frame-Options")
			resp.Header.Del("X-Xss-Protection")
			resp.Header.Del("Set-Cookie")
			resp.Header.Del("Pragma")

			cacheDuration := 14 * 24 * time.Hour
			now := time.Now()

			resp.Header.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cacheDuration/time.Second))
			resp.Header.Set("Expires", now.Add(cacheDuration).Format(http.TimeFormat))
			resp.Header.Set("Last-Modified", now.UTC().Format(http.TimeFormat))

			return nil
		},
	}
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
