package application

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rovilay/url-shortener-service/data"
	"github.com/rovilay/url-shortener-service/handlers"
	"github.com/rs/cors"
)

func (a *App) loadRoutes() {
	router := chi.NewRouter()

	// log start and end of incoming requests
	router.Use(middleware.Logger)

	router.Get("/api", func(w http.ResponseWriter, r *http.Request) {
		var res struct {
			Message string
		}

		res.Message = "Welcome to URL Shortener service"

		msg, err := json.Marshal(res)
		if err != nil {
			fmt.Println("failed to marshall ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(msg)
	})

	router.Route("/api/url", a.loadUrlRoutes)

	a.loadUI(router)

	// CORS configuration
	corsRouter := cors.Default().Handler(router)

	a.router = corsRouter
}

func (a *App) loadUrlRoutes(router chi.Router) {
	urlHandler := handlers.New(&data.RedisRepo{Client: a.db}, a.log)

	router.Group(func(r chi.Router) {
		r.Use(urlHandler.MiddlewareValidateURL)
		r.Post("/shorten", urlHandler.ShortenURL)
	})
	router.Get("/redirect/{hash}", urlHandler.Redirect)
	router.Get("/metrics/shorten", urlHandler.GetShortenMetrics)
	router.Get("/metrics/redirect/{hash}", urlHandler.GetRedirectMetrics)
}

func (a *App) loadUI(router chi.Router) {
	// serve UI
	fs := http.FileServer(http.Dir("./client"))
	router.Handle("/*", fs)
}
