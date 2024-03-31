package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rovilay/url-shortener-service/data"
)

type contextKey string

const UrlCTXKey contextKey = "product"

func (u *URLHandler) MiddlewareValidateURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := &data.URL{}

		// u.log.Debug(r.Body)

		err := url.FromJSON(r.Body)
		if err != nil {
			u.log.Println("[ERROR] deserializing url", err)
			http.Error(w, `{"error": "failed to read url"}`, http.StatusBadRequest)
			return
		}

		// validate url
		err = url.Validate()
		if err != nil {
			u.log.Println("[ERROR] validating URL", err)
			http.Error(
				w, fmt.Sprintf(`{"error": "Error valdating URL: %s"}`, err),
				http.StatusBadRequest,
			)
			return
		}

		// add validated URL
		ctx := context.WithValue(r.Context(), UrlCTXKey, url)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
