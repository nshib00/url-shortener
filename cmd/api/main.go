package main

import (
	"flag"
	"fmt"
	"go-url-shortener/internal/config"
	hDelete "go-url-shortener/internal/http/handlers/urls/delete"
	"go-url-shortener/internal/http/handlers/urls/redirect"
	"go-url-shortener/internal/http/handlers/urls/save"
	"go-url-shortener/internal/http/routers"
	"go-url-shortener/internal/storage/sqlite"
	"go-url-shortener/internal/utils/logger"
	"log"
	"log/slog"
	"net/http"
	"os"
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

	log.Info(fmt.Sprintf("main: starting app [%s]", slog.String("env", cfg.Env)))
	log.Debug("main: debug mode on")

	storage, err := sqlite.NewStorage(cfg.StoragePath)
	if err != nil {
		log.Error(
			"main: failed to init storage",
			logger.Err(err),
			slog.String("storage_path", cfg.StoragePath),
		)
		os.Exit(1)
	}

	saveHandler := save.New(log, storage)
	redirectHandler := redirect.New(log, storage)
	deleteHandler := hDelete.New(log, storage)

	router := routers.New(log, saveHandler, redirectHandler, deleteHandler)

	log.Info(fmt.Sprintf("main: starting server on %s", slog.String("address", cfg.HTTPServer.Address)))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error("main: failed to start server")
	}
	log.Error("server stopped")

}
