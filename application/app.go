package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type App struct {
	router http.Handler
	db     *redis.Client
	config Config
	log    *zerolog.Logger
}

func New(c Config, log *zerolog.Logger) App {
	appLogger := log.With().Str("package", "application").Logger()

	app := App{
		config: c,
		log:    &appLogger,
		db: redis.NewClient(&redis.Options{
			Addr: c.RedisAddress,
		}),
	}

	app.loadRoutes()

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.ServerPort),
		Handler: a.router,
	}

	// ping DB
	err := a.db.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	defer func() {
		if err := a.db.Close(); err != nil {
			a.log.Err(err).Msg("failed to close redis")
		}
	}()

	a.log.Println("Starting Server on port: ", server.Addr)

	ch := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- err
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		// give a bit of time before shutdown
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		return server.Shutdown(timeout)
	}
}
