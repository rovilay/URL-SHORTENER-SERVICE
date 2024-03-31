package handlers

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rovilay/url-shortener-service/data"
	"github.com/rovilay/url-shortener-service/utils"
	"github.com/rs/zerolog"
)

type URLHandler struct {
	DB  *data.RedisRepo
	log *zerolog.Logger
}

func New(db *data.RedisRepo, l *zerolog.Logger) *URLHandler {
	logger := l.With().Str("package", "handlers:URLHandler").Logger()
	return &URLHandler{DB: db, log: &logger}
}

func (u *URLHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	url := r.Context().Value(UrlCTXKey).(*data.URL)

	// validate custom hash if it exist
	if url.ShortHash != "" {
		foundUrl, err := u.DB.FindByHash(r.Context(), url.ShortHash)
		if foundUrl.ShortHash == url.ShortHash {
			// w.WriteHeader(http.StatusNotFound)
			http.Error(w, `{"error": "short code already in use"}`, http.StatusBadRequest)
			return
		} else if err != nil && !errors.Is(err, data.ErrNotExist) {
			u.log.Println("something went wrong: ", err)
			// w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, `{"error": "something went wrong"}`, http.StatusInternalServerError)
			return
		}
	} else {
		url.ShortHash = utils.ShortenURLHash(url.Link)
	}

	url.ID = rand.Int()
	now := time.Now().UTC()
	url.CreatedAt = now.String()

	err := u.DB.Insert(r.Context(), *url)
	if err != nil {
		u.log.Println("failed to insert: ", err)
		http.Error(w, `{"error": "failed to insert"}`, http.StatusBadRequest)
		return
	}

	// count metric
	u.DB.CountMetric(r.Context(), data.Shorten, "")

	w.WriteHeader(http.StatusCreated)

	if err := url.ToJSON(w); err != nil {
		u.log.Println("failed to marshal: ", err)
		http.Error(w, `{"error": "failed to marshal"}`, http.StatusInternalServerError)
		return
	}
}

func (u *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	hashParam := chi.URLParam(r, "hash")

	url, err := u.DB.FindByHash(r.Context(), hashParam)
	if errors.Is(err, data.ErrNotExist) {
		// w.WriteHeader(http.StatusNotFound)
		http.Error(w, `{"error": "url resource not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		u.log.Println("failed to find by id: ", err)
		// w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, `{"error": "something went wrong"}`, http.StatusInternalServerError)
		return
	}

	// count metric
	u.DB.CountMetric(r.Context(), data.Redirect, url.ShortHash)

	http.Redirect(w, r, url.Link, http.StatusSeeOther)
}

type MetricResponse struct {
	Count uint `json:"count"`
}

func (u *URLHandler) GetShortenMetrics(w http.ResponseWriter, r *http.Request) {
	shortenCount, err := u.DB.GetCountMetric(r.Context(), data.Shorten, "")
	if err != nil {
		u.log.Println("something went wrong: ", err)
		// w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, `{"error": "something went wrong"}`, http.StatusInternalServerError)
		return
	}

	res := &MetricResponse{Count: uint(shortenCount)}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		u.log.Println("something went wrong: ", err)
		// w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, `{"error": "something went wrong"}`, http.StatusInternalServerError)
		return
	}
}

func (u *URLHandler) GetRedirectMetrics(w http.ResponseWriter, r *http.Request) {
	hashParam := chi.URLParam(r, "hash")

	redirectCount, err := u.DB.GetCountMetric(r.Context(), data.Redirect, hashParam)
	if err != nil {
		u.log.Println("something went wrong: ", err)
		// w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, `{"error": "something went wrong"}`, http.StatusInternalServerError)
		return
	}

	res := &MetricResponse{Count: uint(redirectCount)}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		u.log.Println("something went wrong: ", err)
		// w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, `{"error": "something went wrong"}`, http.StatusInternalServerError)
		return
	}
}
