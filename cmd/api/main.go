package main

import (
	"flag"
	"fmt"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/config/lib/logger"
	"go-url-shortener/internal/storage/sqlite"
	"log"
	"log/slog"
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

	storage, err := sqlite.NewStorage(cfg.StoragePath)
	if err != nil {
		log.Error("main: failed to init storage", logger.Err(err))
		os.Exit(1)
	}
	_ = storage

	log.Info(fmt.Sprintf("main: starting app [%s]", slog.String("env", cfg.Env)))
	log.Debug("main: debug mode on")
}
