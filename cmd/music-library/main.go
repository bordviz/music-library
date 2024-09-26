package main

import (
	"context"
	"fmt"
	"log/slog"
	"music-library/internal/config"
	libraryhandlers "music-library/internal/handlers/library"
	"music-library/internal/lib/logger/sl"
	mwLogger "music-library/internal/lib/middleware"
	"music-library/internal/logger"
	"music-library/internal/migrations"
	libraryservice "music-library/internal/services/library"
	"music-library/internal/storage/library"
	"music-library/internal/storage/postgresql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	_ "music-library/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

//	@title			Mysic Library Service
//	@version		1.0
//	@description	This is a sample server HTTP server.

// @host		localhost:8080
// @BasePath	/
func main() {
	cfg := config.MustLoad()

	log := logger.New(cfg.Env)
	log.Debug("debug messages are available")
	log.Info("info messages are available")
	log.Warn("warn messages are available")
	log.Error("error messages are available")

	pool, err := postgresql.NewConection(context.TODO(), log, cfg.Database)
	if err != nil {
		log.Error("failed to connect to postgresql", sl.Err(err))
		os.Exit(1)
	}

	if err := migrations.CreateMigrations(cfg, "up"); err != nil {
		log.Error("failed to create migrations", sl.Err(err))
		os.Exit(1)
	}
	log.Info("migrations applied successfully")

	libraryDB := library.NewLibraryDB(log)
	libraryService := libraryservice.NewLibraryService(log, pool, libraryDB, cfg.LibraryServer)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(mwLogger.New(log))
	log.Info("middleware successfully conected")

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	log.Info("cors successfully conected")

	router.Route("/", libraryhandlers.AddHandler(context.TODO(), log, libraryService))

	router.Mount("/swagger", httpSwagger.WrapHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPServer.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		log.Info("starting server", slog.String("addr", fmt.Sprintf("::%d", cfg.HTTPServer.Port)))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to listen and serve", sl.Err(err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	stopSignal := <-stop
	log.Info("stoppping server", slog.String("signal", stopSignal.String()))
	ctx, close := context.WithTimeout(context.Background(), time.Minute)
	defer close()
	srv.Shutdown(ctx)
	pool.Close()
	log.Info("server was stopped")
}
