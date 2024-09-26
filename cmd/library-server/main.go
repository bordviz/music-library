package main

import (
	"context"
	"fmt"
	"music-library/internal/config"
	"music-library/internal/domain/dto"
	"music-library/internal/handlers"
	mwLogger "music-library/internal/lib/middleware"
	"music-library/internal/logger"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const text = `Ooh baby, don't you know I suffer?
Ooh baby, can you hear me moan?
You caught me under false pretenses
How long before you let me go?

Ooh
You set my soul alight
Ooh
You set my soul alight

Glaciers melting in the dead of night (ooh)
And the superstars sucked into the supermassive (you set my soul alight)
Glaciers melting in the dead of night
And the superstars sucked into the (you set my soul)
(Into the supermassive)

I thought I was a fool for no one
Ooh baby, I'm a fool for you
You're the queen of the superficial
And how long before you tell the truth?

Ooh
You set my soul alight
Ooh
You set my soul alight

Glaciers melting in the dead of night (ooh)
And the superstars sucked into the supermassive (you set my soul alight)
Glaciers melting in the dead of night
And the superstars sucked into the (you set my soul)
(Into the supermassive)

Supermassive black hole
Supermassive black hole
Supermassive black hole
Supermassive black hole

Glaciers melting in the dead of night
And the superstars sucked into the supermassive
Glaciers melting in the dead of night
And the superstars sucked into the supermassive
Glaciers melting in the dead of night (ooh)
And the superstars sucked into the supermassive (you set my soul alight)
Glaciers melting in the dead of night
And the superstars sucked into the (you set my soul)
(Into the supermassive)

Supermassive black hole
Supermassive black hole
Supermassive black hole
Supermassive black hole`

func main() {
	cfg := config.MustLoad()
	log := logger.New(cfg.Env)

	model := dto.Song{
		Group:       "Muse",
		Song:        "Supermassive Black Hole",
		Text:        text,
		ReleaseDate: "16.07.2006",
		Patronymic:  "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(mwLogger.New(log))

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Get("/info", func(w http.ResponseWriter, r *http.Request) {
		group := r.URL.Query().Get("group")
		if group == "" {
			handlers.ErrorResponse(w, r, 400, "group parameter is required")
			return
		}

		song := r.URL.Query().Get("song")
		if song == "" {
			handlers.ErrorResponse(w, r, 400, "song parameter is required")
			return
		}

		if strings.EqualFold(group, model.Group) && strings.EqualFold(song, model.Song) {
			handlers.SuccessResponse(w, r, 200, model)
			return
		}

		handlers.ErrorResponse(w, r, 404, "song not found")
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.LibraryServer.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	stopSignal := <-stop
	ctx, close := context.WithTimeout(context.Background(), time.Minute)
	defer close()
	srv.Shutdown(ctx)
	fmt.Println("Shutting down, signal:", stopSignal.String())
}
