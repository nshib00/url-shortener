package main

import (
	"flag"
	"fmt"
	"go-url-shortener/internal/config"
	urlHandlers "go-url-shortener/internal/http/handlers/url"
	httpMiddleware "go-url-shortener/internal/http/middleware"
	"go-url-shortener/internal/storage/sqlite"
	"go-url-shortener/internal/utils/logger"
	"log"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func setupLogger(envType config.EnvType) *slog.Logger {
	switch envType {
	case config.EnvLocal:
		return slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case config.EnvProd:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return nil
}

func main() {
	cfgTypeArg := flag.String("config", "local", "Тип config-файла. Опции: local, prod.")
	flag.Parse()

	envType, err := config.ParseEnvType(*cfgTypeArg)
	if err != nil {
		log.Fatal("main: wrong env type")
	}
	cfg := config.MustLoad(envType)
	log := setupLogger(envType)

	storage, err := sqlite.NewStorage(cfg.StoragePath)
	if err != nil {
		log.Error("main: failed to init storage", logger.Err(err))
		os.Exit(1)
	}
	_ = storage

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(httpMiddleware.LoggerMiddleware(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", urlHandlers.NewSaveHandler(log))

	log.Info(fmt.Sprintf("main: starting app [%s]", slog.String("env", cfg.Env)))
	log.Debug("main: debug mode on")
}
