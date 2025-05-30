package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (s *server) handleTTS(w http.ResponseWriter, r *http.Request) error {
	cacheKey := func(url *url.URL) string {
		return fmt.Sprintf("%s:%s?%s", "tts", url.EscapedPath(), url.Query().Encode())
	}

	cached, err := s.cache.Get(cacheKey(r.URL))
	if err == nil && len(cached) > 0 {
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Write(cached)

		return nil
	}

	// "Free" TTS solution with good support for danish
	// https://translate.google.com/translate_tts?ie=UTF-8&client=tw-ob&q=test+af+tekst+til+tale&tl=da

	req, err := http.NewRequest("GET", "https://translate.google.com/translate_tts?ie=UTF-8&client=tw-ob&tl=da", nil)
	if err != nil {
		return fmt.Errorf("http.NewRequest: %w", err)
	}

	if req.URL.RawQuery == "" || r.URL.RawQuery == "" {
		req.URL.RawQuery = req.URL.RawQuery + r.URL.RawQuery
	} else {
		req.URL.RawQuery = req.URL.RawQuery + "&" + r.URL.RawQuery
	}

	req.Header.Set("User-Agent", r.Header.Get("User-Agent"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.DefaultClient.Do: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	if err := s.cache.Set(cacheKey(r.URL), body); err != nil {
		return fmt.Errorf("memcache.Set: %w", err)
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Write(body)

	return nil
}
